package disk

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
)

// Service handles GCP Compute disk ingestion.
type Service struct {
	client  *Client
	db      *gorm.DB
	history *HistoryService
}

// NewService creates a new disk ingestion service.
func NewService(client *Client, db *gorm.DB) *Service {
	return &Service{
		client:  client,
		db:      db,
		history: NewHistoryService(db),
	}
}

// IngestParams contains parameters for disk ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of disk ingestion.
type IngestResult struct {
	ProjectID      string
	DiskCount      int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches disks from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch disks from GCP
	disks, err := s.client.ListDisks(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list disks: %w", err)
	}

	// Convert to bronze models
	bronzeDisks := make([]bronze.GCPComputeDisk, 0, len(disks))
	for _, d := range disks {
		bronzeDisks = append(bronzeDisks, ConvertDisk(d, params.ProjectID, collectedAt))
	}

	// Save to database
	if err := s.saveDisks(ctx, bronzeDisks); err != nil {
		return nil, fmt.Errorf("failed to save disks: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		DiskCount:      len(bronzeDisks),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveDisks saves disks to the database with history tracking.
func (s *Service) saveDisks(ctx context.Context, disks []bronze.GCPComputeDisk) error {
	if len(disks) == 0 {
		return nil
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, disk := range disks {
			// Load existing disk with all relations
			var existing *bronze.GCPComputeDisk
			var old bronze.GCPComputeDisk
			err := tx.Preload("Labels").Preload("Licenses").
				Where("resource_id = ?", disk.ResourceID).
				First(&old).Error
			if err == nil {
				existing = &old
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to load existing disk %s: %w", disk.Name, err)
			}

			// Compute diff
			diff := DiffDisk(existing, &disk)

			// Skip if no changes
			if !diff.HasAnyChange() && existing != nil {
				// Update collected_at only
				if err := tx.Model(&bronze.GCPComputeDisk{}).
					Where("resource_id = ?", disk.ResourceID).
					Update("collected_at", disk.CollectedAt).Error; err != nil {
					return fmt.Errorf("failed to update collected_at for disk %s: %w", disk.Name, err)
				}
				continue
			}

			// Delete old relations (manual cascade)
			if existing != nil {
				if err := s.deleteDiskRelations(tx, disk.ResourceID); err != nil {
					return fmt.Errorf("failed to delete old relations for disk %s: %w", disk.Name, err)
				}
			}

			// Upsert disk
			if err := tx.Save(&disk).Error; err != nil {
				return fmt.Errorf("failed to upsert disk %s: %w", disk.Name, err)
			}

			// Create new relations
			if err := s.createDiskRelations(tx, disk.ResourceID, &disk); err != nil {
				return fmt.Errorf("failed to create relations for disk %s: %w", disk.Name, err)
			}

			// Track history
			if diff.IsNew {
				if err := s.history.CreateHistory(tx, &disk, now); err != nil {
					return fmt.Errorf("failed to create history for disk %s: %w", disk.Name, err)
				}
			} else {
				if err := s.history.UpdateHistory(tx, existing, &disk, diff, now); err != nil {
					return fmt.Errorf("failed to update history for disk %s: %w", disk.Name, err)
				}
			}
		}

		return nil
	})
}

// deleteDiskRelations deletes all related records for a disk.
func (s *Service) deleteDiskRelations(tx *gorm.DB, diskResourceID string) error {
	// Delete labels
	if err := tx.Where("disk_resource_id = ?", diskResourceID).Delete(&bronze.GCPComputeDiskLabel{}).Error; err != nil {
		return err
	}

	// Delete licenses
	if err := tx.Where("disk_resource_id = ?", diskResourceID).Delete(&bronze.GCPComputeDiskLicense{}).Error; err != nil {
		return err
	}

	return nil
}

// createDiskRelations creates all related records for a disk.
func (s *Service) createDiskRelations(tx *gorm.DB, diskResourceID string, disk *bronze.GCPComputeDisk) error {
	// Create labels
	for i := range disk.Labels {
		disk.Labels[i].DiskResourceID = diskResourceID
	}
	if len(disk.Labels) > 0 {
		if err := tx.Create(&disk.Labels).Error; err != nil {
			return fmt.Errorf("failed to create labels: %w", err)
		}
	}

	// Create licenses
	for i := range disk.Licenses {
		disk.Licenses[i].DiskResourceID = diskResourceID
	}
	if len(disk.Licenses) > 0 {
		if err := tx.Create(&disk.Licenses).Error; err != nil {
			return fmt.Errorf("failed to create licenses: %w", err)
		}
	}

	return nil
}

// DeleteStaleDisks removes disks that were not collected in the latest run.
// Also closes history records for deleted disks.
func (s *Service) DeleteStaleDisks(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find stale disks
		var staleDisks []bronze.GCPComputeDisk
		if err := tx.Where("project_id = ? AND collected_at < ?", projectID, collectedAt).
			Find(&staleDisks).Error; err != nil {
			return err
		}

		// Close history and delete each stale disk
		for _, d := range staleDisks {
			// Close history
			if err := s.history.CloseHistory(tx, d.ResourceID, now); err != nil {
				return fmt.Errorf("failed to close history for disk %s: %w", d.ResourceID, err)
			}

			// Delete relations
			if err := s.deleteDiskRelations(tx, d.ResourceID); err != nil {
				return fmt.Errorf("failed to delete relations for disk %s: %w", d.ResourceID, err)
			}

			// Delete disk
			if err := tx.Delete(&d).Error; err != nil {
				return fmt.Errorf("failed to delete disk %s: %w", d.ResourceID, err)
			}
		}

		return nil
	})
}
