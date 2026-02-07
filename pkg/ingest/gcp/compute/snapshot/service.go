package snapshot

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
)

// Service handles GCP Compute snapshot ingestion.
type Service struct {
	client  *Client
	db      *gorm.DB
	history *HistoryService
}

// NewService creates a new snapshot ingestion service.
func NewService(client *Client, db *gorm.DB) *Service {
	return &Service{
		client:  client,
		db:      db,
		history: NewHistoryService(db),
	}
}

// IngestParams contains parameters for snapshot ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of snapshot ingestion.
type IngestResult struct {
	ProjectID      string
	SnapshotCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches snapshots from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch snapshots from GCP
	snapshots, err := s.client.ListSnapshots(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}

	// Convert to bronze models
	bronzeSnapshots := make([]bronze.GCPComputeSnapshot, 0, len(snapshots))
	for _, snap := range snapshots {
		bs, err := ConvertSnapshot(snap, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert snapshot: %w", err)
		}
		bronzeSnapshots = append(bronzeSnapshots, bs)
	}

	// Save to database
	if err := s.saveSnapshots(ctx, bronzeSnapshots); err != nil {
		return nil, fmt.Errorf("failed to save snapshots: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		SnapshotCount:  len(bronzeSnapshots),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveSnapshots saves snapshots to the database with history tracking.
func (s *Service) saveSnapshots(ctx context.Context, snapshots []bronze.GCPComputeSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, snap := range snapshots {
			// Load existing snapshot with all relations
			var existing *bronze.GCPComputeSnapshot
			var old bronze.GCPComputeSnapshot
			err := tx.Preload("Labels").Preload("Licenses").
				Where("resource_id = ?", snap.ResourceID).
				First(&old).Error
			if err == nil {
				existing = &old
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to load existing snapshot %s: %w", snap.Name, err)
			}

			// Compute diff
			diff := DiffSnapshot(existing, &snap)

			// Skip if no changes
			if !diff.HasAnyChange() && existing != nil {
				// Update collected_at only
				if err := tx.Model(&bronze.GCPComputeSnapshot{}).
					Where("resource_id = ?", snap.ResourceID).
					Update("collected_at", snap.CollectedAt).Error; err != nil {
					return fmt.Errorf("failed to update collected_at for snapshot %s: %w", snap.Name, err)
				}
				continue
			}

			// Delete old relations (manual cascade)
			if existing != nil {
				if err := s.deleteSnapshotRelations(tx, snap.ResourceID); err != nil {
					return fmt.Errorf("failed to delete old relations for snapshot %s: %w", snap.Name, err)
				}
			}

			// Upsert snapshot
			if err := tx.Save(&snap).Error; err != nil {
				return fmt.Errorf("failed to upsert snapshot %s: %w", snap.Name, err)
			}

			// Create new relations
			if err := s.createSnapshotRelations(tx, snap.ResourceID, &snap); err != nil {
				return fmt.Errorf("failed to create relations for snapshot %s: %w", snap.Name, err)
			}

			// Track history
			if diff.IsNew {
				if err := s.history.CreateHistory(tx, &snap, now); err != nil {
					return fmt.Errorf("failed to create history for snapshot %s: %w", snap.Name, err)
				}
			} else {
				if err := s.history.UpdateHistory(tx, existing, &snap, diff, now); err != nil {
					return fmt.Errorf("failed to update history for snapshot %s: %w", snap.Name, err)
				}
			}
		}

		return nil
	})
}

// deleteSnapshotRelations deletes all related records for a snapshot.
func (s *Service) deleteSnapshotRelations(tx *gorm.DB, snapshotResourceID string) error {
	// Delete labels
	if err := tx.Where("snapshot_resource_id = ?", snapshotResourceID).Delete(&bronze.GCPComputeSnapshotLabel{}).Error; err != nil {
		return err
	}

	// Delete licenses
	if err := tx.Where("snapshot_resource_id = ?", snapshotResourceID).Delete(&bronze.GCPComputeSnapshotLicense{}).Error; err != nil {
		return err
	}

	return nil
}

// createSnapshotRelations creates all related records for a snapshot.
func (s *Service) createSnapshotRelations(tx *gorm.DB, snapshotResourceID string, snap *bronze.GCPComputeSnapshot) error {
	// Create labels
	for i := range snap.Labels {
		snap.Labels[i].SnapshotResourceID = snapshotResourceID
	}
	if len(snap.Labels) > 0 {
		if err := tx.Create(&snap.Labels).Error; err != nil {
			return fmt.Errorf("failed to create labels: %w", err)
		}
	}

	// Create licenses
	for i := range snap.Licenses {
		snap.Licenses[i].SnapshotResourceID = snapshotResourceID
	}
	if len(snap.Licenses) > 0 {
		if err := tx.Create(&snap.Licenses).Error; err != nil {
			return fmt.Errorf("failed to create licenses: %w", err)
		}
	}

	return nil
}

// DeleteStaleSnapshots removes snapshots that were not collected in the latest run.
// Also closes history records for deleted snapshots.
func (s *Service) DeleteStaleSnapshots(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find stale snapshots
		var staleSnapshots []bronze.GCPComputeSnapshot
		if err := tx.Where("project_id = ? AND collected_at < ?", projectID, collectedAt).
			Find(&staleSnapshots).Error; err != nil {
			return err
		}

		// Close history and delete each stale snapshot
		for _, snap := range staleSnapshots {
			// Close history
			if err := s.history.CloseHistory(tx, snap.ResourceID, now); err != nil {
				return fmt.Errorf("failed to close history for snapshot %s: %w", snap.ResourceID, err)
			}

			// Delete relations
			if err := s.deleteSnapshotRelations(tx, snap.ResourceID); err != nil {
				return fmt.Errorf("failed to delete relations for snapshot %s: %w", snap.ResourceID, err)
			}

			// Delete snapshot
			if err := tx.Delete(&snap).Error; err != nil {
				return fmt.Errorf("failed to delete snapshot %s: %w", snap.ResourceID, err)
			}
		}

		return nil
	})
}
