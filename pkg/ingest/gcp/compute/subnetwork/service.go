package subnetwork

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
)

// Service handles GCP Compute subnetwork ingestion.
type Service struct {
	client  *Client
	db      *gorm.DB
	history *HistoryService
}

// NewService creates a new subnetwork ingestion service.
func NewService(client *Client, db *gorm.DB) *Service {
	return &Service{
		client:  client,
		db:      db,
		history: NewHistoryService(db),
	}
}

// IngestParams contains parameters for subnetwork ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of subnetwork ingestion.
type IngestResult struct {
	ProjectID       string
	SubnetworkCount int
	CollectedAt     time.Time
	DurationMillis  int64
}

// Ingest fetches subnetworks from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch subnetworks from GCP
	subnetworks, err := s.client.ListSubnetworks(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list subnetworks: %w", err)
	}

	// Convert to bronze models
	bronzeSubnetworks := make([]bronze.GCPComputeSubnetwork, 0, len(subnetworks))
	for _, sn := range subnetworks {
		bronzeSubnetworks = append(bronzeSubnetworks, ConvertSubnetwork(sn, params.ProjectID, collectedAt))
	}

	// Save to database
	if err := s.saveSubnetworks(ctx, bronzeSubnetworks); err != nil {
		return nil, fmt.Errorf("failed to save subnetworks: %w", err)
	}

	return &IngestResult{
		ProjectID:       params.ProjectID,
		SubnetworkCount: len(bronzeSubnetworks),
		CollectedAt:     collectedAt,
		DurationMillis:  time.Since(startTime).Milliseconds(),
	}, nil
}

// saveSubnetworks saves subnetworks to the database with history tracking.
func (s *Service) saveSubnetworks(ctx context.Context, subnetworks []bronze.GCPComputeSubnetwork) error {
	if len(subnetworks) == 0 {
		return nil
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, subnet := range subnetworks {
			// Load existing subnetwork with all relations
			var existing *bronze.GCPComputeSubnetwork
			var old bronze.GCPComputeSubnetwork
			err := tx.Preload("SecondaryIpRanges").
				Where("resource_id = ?", subnet.ResourceID).
				First(&old).Error
			if err == nil {
				existing = &old
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to load existing subnetwork %s: %w", subnet.Name, err)
			}

			// Compute diff
			diff := DiffSubnetwork(existing, &subnet)

			// Skip if no changes
			if !diff.HasAnyChange() && existing != nil {
				// Update collected_at only
				if err := tx.Model(&bronze.GCPComputeSubnetwork{}).
					Where("resource_id = ?", subnet.ResourceID).
					Update("collected_at", subnet.CollectedAt).Error; err != nil {
					return fmt.Errorf("failed to update collected_at for subnetwork %s: %w", subnet.Name, err)
				}
				continue
			}

			// Delete old relations (manual cascade)
			if existing != nil {
				if err := s.deleteSubnetworkRelations(tx, subnet.ResourceID); err != nil {
					return fmt.Errorf("failed to delete old relations for subnetwork %s: %w", subnet.Name, err)
				}
			}

			// Upsert subnetwork
			if err := tx.Save(&subnet).Error; err != nil {
				return fmt.Errorf("failed to upsert subnetwork %s: %w", subnet.Name, err)
			}

			// Create new relations
			if err := s.createSubnetworkRelations(tx, subnet.ResourceID, &subnet); err != nil {
				return fmt.Errorf("failed to create relations for subnetwork %s: %w", subnet.Name, err)
			}

			// Track history
			if diff.IsNew {
				if err := s.history.CreateHistory(tx, &subnet, now); err != nil {
					return fmt.Errorf("failed to create history for subnetwork %s: %w", subnet.Name, err)
				}
			} else {
				if err := s.history.UpdateHistory(tx, existing, &subnet, diff, now); err != nil {
					return fmt.Errorf("failed to update history for subnetwork %s: %w", subnet.Name, err)
				}
			}
		}

		return nil
	})
}

// deleteSubnetworkRelations deletes all related records for a subnetwork.
func (s *Service) deleteSubnetworkRelations(tx *gorm.DB, subnetworkResourceID string) error {
	// Delete secondary ranges
	if err := tx.Where("subnetwork_resource_id = ?", subnetworkResourceID).Delete(&bronze.GCPComputeSubnetworkSecondaryRange{}).Error; err != nil {
		return err
	}

	return nil
}

// createSubnetworkRelations creates all related records for a subnetwork.
func (s *Service) createSubnetworkRelations(tx *gorm.DB, subnetworkResourceID string, subnet *bronze.GCPComputeSubnetwork) error {
	// Create secondary ranges
	for i := range subnet.SecondaryIpRanges {
		subnet.SecondaryIpRanges[i].SubnetworkResourceID = subnetworkResourceID
	}
	if len(subnet.SecondaryIpRanges) > 0 {
		if err := tx.Create(&subnet.SecondaryIpRanges).Error; err != nil {
			return fmt.Errorf("failed to create secondary ranges: %w", err)
		}
	}

	return nil
}

// DeleteStaleSubnetworks removes subnetworks that were not collected in the latest run.
// Also closes history records for deleted subnetworks.
func (s *Service) DeleteStaleSubnetworks(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find stale subnetworks
		var staleSubnetworks []bronze.GCPComputeSubnetwork
		if err := tx.Where("project_id = ? AND collected_at < ?", projectID, collectedAt).
			Find(&staleSubnetworks).Error; err != nil {
			return err
		}

		// Close history and delete each stale subnetwork
		for _, sn := range staleSubnetworks {
			// Close history
			if err := s.history.CloseHistory(tx, sn.ResourceID, now); err != nil {
				return fmt.Errorf("failed to close history for subnetwork %s: %w", sn.ResourceID, err)
			}

			// Delete relations
			if err := s.deleteSubnetworkRelations(tx, sn.ResourceID); err != nil {
				return fmt.Errorf("failed to delete relations for subnetwork %s: %w", sn.ResourceID, err)
			}

			// Delete subnetwork
			if err := tx.Delete(&sn).Error; err != nil {
				return fmt.Errorf("failed to delete subnetwork %s: %w", sn.ResourceID, err)
			}
		}

		return nil
	})
}
