package targetinstance

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
)

// Service handles GCP Compute target instance ingestion.
type Service struct {
	client  *Client
	db      *gorm.DB
	history *HistoryService
}

// NewService creates a new target instance ingestion service.
func NewService(client *Client, db *gorm.DB) *Service {
	return &Service{
		client:  client,
		db:      db,
		history: NewHistoryService(db),
	}
}

// IngestParams contains parameters for target instance ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of target instance ingestion.
type IngestResult struct {
	ProjectID           string
	TargetInstanceCount int
	CollectedAt         time.Time
	DurationMillis      int64
}

// Ingest fetches target instances from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch target instances from GCP
	targetInstances, err := s.client.ListTargetInstances(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list target instances: %w", err)
	}

	// Convert to bronze models
	bronzeTargetInstances := make([]bronze.GCPComputeTargetInstance, 0, len(targetInstances))
	for _, ti := range targetInstances {
		bronzeTargetInstances = append(bronzeTargetInstances, ConvertTargetInstance(ti, params.ProjectID, collectedAt))
	}

	// Save to database
	if err := s.saveTargetInstances(ctx, bronzeTargetInstances); err != nil {
		return nil, fmt.Errorf("failed to save target instances: %w", err)
	}

	return &IngestResult{
		ProjectID:           params.ProjectID,
		TargetInstanceCount: len(bronzeTargetInstances),
		CollectedAt:         collectedAt,
		DurationMillis:      time.Since(startTime).Milliseconds(),
	}, nil
}

// saveTargetInstances saves target instances to the database with history tracking.
func (s *Service) saveTargetInstances(ctx context.Context, targetInstances []bronze.GCPComputeTargetInstance) error {
	if len(targetInstances) == 0 {
		return nil
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, ti := range targetInstances {
			// Load existing target instance
			var existing *bronze.GCPComputeTargetInstance
			var old bronze.GCPComputeTargetInstance
			err := tx.Where("resource_id = ?", ti.ResourceID).
				First(&old).Error
			if err == nil {
				existing = &old
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to load existing target instance %s: %w", ti.Name, err)
			}

			// Compute diff
			diff := DiffTargetInstance(existing, &ti)

			// Skip if no changes
			if !diff.HasAnyChange() && existing != nil {
				// Update collected_at only
				if err := tx.Model(&bronze.GCPComputeTargetInstance{}).
					Where("resource_id = ?", ti.ResourceID).
					Update("collected_at", ti.CollectedAt).Error; err != nil {
					return fmt.Errorf("failed to update collected_at for target instance %s: %w", ti.Name, err)
				}
				continue
			}

			// Upsert target instance
			if err := tx.Save(&ti).Error; err != nil {
				return fmt.Errorf("failed to upsert target instance %s: %w", ti.Name, err)
			}

			// Track history
			if diff.IsNew {
				if err := s.history.CreateHistory(tx, &ti, now); err != nil {
					return fmt.Errorf("failed to create history for target instance %s: %w", ti.Name, err)
				}
			} else {
				if err := s.history.UpdateHistory(tx, existing, &ti, diff, now); err != nil {
					return fmt.Errorf("failed to update history for target instance %s: %w", ti.Name, err)
				}
			}
		}

		return nil
	})
}

// DeleteStaleTargetInstances removes target instances that were not collected in the latest run.
// Also closes history records for deleted target instances.
func (s *Service) DeleteStaleTargetInstances(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find stale target instances
		var staleTargetInstances []bronze.GCPComputeTargetInstance
		if err := tx.Where("project_id = ? AND collected_at < ?", projectID, collectedAt).
			Find(&staleTargetInstances).Error; err != nil {
			return err
		}

		// Close history and delete each stale target instance
		for _, ti := range staleTargetInstances {
			// Close history
			if err := s.history.CloseHistory(tx, ti.ResourceID, now); err != nil {
				return fmt.Errorf("failed to close history for target instance %s: %w", ti.ResourceID, err)
			}

			// Delete target instance
			if err := tx.Delete(&ti).Error; err != nil {
				return fmt.Errorf("failed to delete target instance %s: %w", ti.ResourceID, err)
			}
		}

		return nil
	})
}
