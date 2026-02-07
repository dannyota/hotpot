package connector

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
)

// Service handles GCP VPC Access connector ingestion.
type Service struct {
	client  *Client
	db      *gorm.DB
	history *HistoryService
}

// NewService creates a new connector ingestion service.
func NewService(client *Client, db *gorm.DB) *Service {
	return &Service{
		client:  client,
		db:      db,
		history: NewHistoryService(db),
	}
}

// IngestParams contains parameters for connector ingestion.
type IngestParams struct {
	ProjectID string
	Regions   []string
}

// IngestResult contains the result of connector ingestion.
type IngestResult struct {
	ProjectID      string
	ConnectorCount int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches connectors from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch connectors from GCP across regions
	connectors, err := s.client.ListConnectors(ctx, params.ProjectID, params.Regions)
	if err != nil {
		return nil, fmt.Errorf("failed to list connectors: %w", err)
	}

	// Convert to bronze models
	bronzeConnectors := make([]bronze.GCPVpcAccessConnector, 0, len(connectors))
	for _, c := range connectors {
		bc, err := ConvertConnector(c, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert connector: %w", err)
		}
		bronzeConnectors = append(bronzeConnectors, bc)
	}

	// Save to database
	if err := s.saveConnectors(ctx, bronzeConnectors); err != nil {
		return nil, fmt.Errorf("failed to save connectors: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		ConnectorCount: len(bronzeConnectors),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveConnectors saves connectors to the database with history tracking.
func (s *Service) saveConnectors(ctx context.Context, connectors []bronze.GCPVpcAccessConnector) error {
	if len(connectors) == 0 {
		return nil
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, c := range connectors {
			// Load existing connector
			var existing *bronze.GCPVpcAccessConnector
			var old bronze.GCPVpcAccessConnector
			err := tx.Where("resource_id = ?", c.ResourceID).First(&old).Error
			if err == nil {
				existing = &old
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to load existing connector %s: %w", c.ResourceID, err)
			}

			// Compute diff
			diff := DiffConnector(existing, &c)

			// Skip if no changes
			if !diff.HasAnyChange() && existing != nil {
				// Update collected_at only
				if err := tx.Model(&bronze.GCPVpcAccessConnector{}).
					Where("resource_id = ?", c.ResourceID).
					Update("collected_at", c.CollectedAt).Error; err != nil {
					return fmt.Errorf("failed to update collected_at for connector %s: %w", c.ResourceID, err)
				}
				continue
			}

			// Upsert connector
			if err := tx.Save(&c).Error; err != nil {
				return fmt.Errorf("failed to upsert connector %s: %w", c.ResourceID, err)
			}

			// Track history
			if diff.IsNew {
				if err := s.history.CreateHistory(tx, &c, now); err != nil {
					return fmt.Errorf("failed to create history for connector %s: %w", c.ResourceID, err)
				}
			} else {
				if err := s.history.UpdateHistory(tx, existing, &c, diff, now); err != nil {
					return fmt.Errorf("failed to update history for connector %s: %w", c.ResourceID, err)
				}
			}
		}

		return nil
	})
}

// DeleteStaleConnectors removes connectors that were not collected in the latest run.
// Also closes history records for deleted connectors.
func (s *Service) DeleteStaleConnectors(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find stale connectors
		var stale []bronze.GCPVpcAccessConnector
		if err := tx.Where("project_id = ? AND collected_at < ?", projectID, collectedAt).
			Find(&stale).Error; err != nil {
			return err
		}

		// Close history and delete each stale connector
		for _, c := range stale {
			if err := s.history.CloseHistory(tx, c.ResourceID, now); err != nil {
				return fmt.Errorf("failed to close history for connector %s: %w", c.ResourceID, err)
			}
			if err := tx.Delete(&c).Error; err != nil {
				return fmt.Errorf("failed to delete connector %s: %w", c.ResourceID, err)
			}
		}

		return nil
	})
}
