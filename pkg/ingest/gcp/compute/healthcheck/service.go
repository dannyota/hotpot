package healthcheck

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
)

// Service handles GCP Compute health check ingestion.
type Service struct {
	client  *Client
	db      *gorm.DB
	history *HistoryService
}

// NewService creates a new health check ingestion service.
func NewService(client *Client, db *gorm.DB) *Service {
	return &Service{
		client:  client,
		db:      db,
		history: NewHistoryService(db),
	}
}

// IngestParams contains parameters for health check ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of health check ingestion.
type IngestResult struct {
	ProjectID        string
	HealthCheckCount int
	CollectedAt      time.Time
	DurationMillis   int64
}

// Ingest fetches health checks from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch health checks from GCP
	healthChecks, err := s.client.ListHealthChecks(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list health checks: %w", err)
	}

	// Convert to bronze models
	bronzeChecks := make([]bronze.GCPComputeHealthCheck, 0, len(healthChecks))
	for _, hc := range healthChecks {
		bc, err := ConvertHealthCheck(hc, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert health check: %w", err)
		}
		bronzeChecks = append(bronzeChecks, bc)
	}

	// Save to database
	if err := s.saveHealthChecks(ctx, bronzeChecks); err != nil {
		return nil, fmt.Errorf("failed to save health checks: %w", err)
	}

	return &IngestResult{
		ProjectID:        params.ProjectID,
		HealthCheckCount: len(bronzeChecks),
		CollectedAt:      collectedAt,
		DurationMillis:   time.Since(startTime).Milliseconds(),
	}, nil
}

// saveHealthChecks saves health checks to the database with history tracking.
func (s *Service) saveHealthChecks(ctx context.Context, checks []bronze.GCPComputeHealthCheck) error {
	if len(checks) == 0 {
		return nil
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, check := range checks {
			// Load existing health check (no Preload — no child tables)
			var existing *bronze.GCPComputeHealthCheck
			var old bronze.GCPComputeHealthCheck
			err := tx.Where("resource_id = ?", check.ResourceID).
				First(&old).Error
			if err == nil {
				existing = &old
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to load existing health check %s: %w", check.Name, err)
			}

			// Compute diff
			diff := DiffHealthCheck(existing, &check)

			// Skip if no changes
			if !diff.HasAnyChange() && existing != nil {
				// Update collected_at only
				if err := tx.Model(&bronze.GCPComputeHealthCheck{}).
					Where("resource_id = ?", check.ResourceID).
					Update("collected_at", check.CollectedAt).Error; err != nil {
					return fmt.Errorf("failed to update collected_at for health check %s: %w", check.Name, err)
				}
				continue
			}

			// Upsert health check (no relations to delete/create)
			if err := tx.Save(&check).Error; err != nil {
				return fmt.Errorf("failed to upsert health check %s: %w", check.Name, err)
			}

			// Track history
			if diff.IsNew {
				if err := s.history.CreateHistory(tx, &check, now); err != nil {
					return fmt.Errorf("failed to create history for health check %s: %w", check.Name, err)
				}
			} else {
				if err := s.history.UpdateHistory(tx, existing, &check, diff, now); err != nil {
					return fmt.Errorf("failed to update history for health check %s: %w", check.Name, err)
				}
			}
		}

		return nil
	})
}

// DeleteStaleHealthChecks removes health checks that were not collected in the latest run.
// Also closes history records for deleted health checks.
func (s *Service) DeleteStaleHealthChecks(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find stale health checks
		var staleChecks []bronze.GCPComputeHealthCheck
		if err := tx.Where("project_id = ? AND collected_at < ?", projectID, collectedAt).
			Find(&staleChecks).Error; err != nil {
			return err
		}

		// Close history and delete each stale health check
		for _, check := range staleChecks {
			// Close history
			if err := s.history.CloseHistory(tx, check.ResourceID, now); err != nil {
				return fmt.Errorf("failed to close history for health check %s: %w", check.ResourceID, err)
			}

			// Delete health check (no relations to delete)
			if err := tx.Delete(&check).Error; err != nil {
				return fmt.Errorf("failed to delete health check %s: %w", check.ResourceID, err)
			}
		}

		return nil
	})
}
