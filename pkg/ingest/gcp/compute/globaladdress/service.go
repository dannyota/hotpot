package globaladdress

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
)

// Service handles GCP Compute global address ingestion.
type Service struct {
	client  *Client
	db      *gorm.DB
	history *HistoryService
}

// NewService creates a new global address ingestion service.
func NewService(client *Client, db *gorm.DB) *Service {
	return &Service{
		client:  client,
		db:      db,
		history: NewHistoryService(db),
	}
}

// IngestParams contains parameters for global address ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of global address ingestion.
type IngestResult struct {
	ProjectID          string
	GlobalAddressCount int
	CollectedAt        time.Time
	DurationMillis     int64
}

// Ingest fetches global addresses from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch global addresses from GCP
	addresses, err := s.client.ListGlobalAddresses(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list global addresses: %w", err)
	}

	// Convert to bronze models
	bronzeAddresses := make([]bronze.GCPComputeGlobalAddress, 0, len(addresses))
	for _, a := range addresses {
		addr, err := ConvertGlobalAddress(a, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert global address: %w", err)
		}
		bronzeAddresses = append(bronzeAddresses, addr)
	}

	// Save to database
	if err := s.saveGlobalAddresses(ctx, bronzeAddresses); err != nil {
		return nil, fmt.Errorf("failed to save global addresses: %w", err)
	}

	return &IngestResult{
		ProjectID:          params.ProjectID,
		GlobalAddressCount: len(bronzeAddresses),
		CollectedAt:        collectedAt,
		DurationMillis:     time.Since(startTime).Milliseconds(),
	}, nil
}

// saveGlobalAddresses saves global addresses to the database with history tracking.
func (s *Service) saveGlobalAddresses(ctx context.Context, addresses []bronze.GCPComputeGlobalAddress) error {
	if len(addresses) == 0 {
		return nil
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, addr := range addresses {
			// Load existing address with all relations
			var existing *bronze.GCPComputeGlobalAddress
			var old bronze.GCPComputeGlobalAddress
			err := tx.Preload("Labels").
				Where("resource_id = ?", addr.ResourceID).
				First(&old).Error
			if err == nil {
				existing = &old
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to load existing global address %s: %w", addr.Name, err)
			}

			// Compute diff
			diff := DiffGlobalAddress(existing, &addr)

			// Skip if no changes
			if !diff.HasAnyChange() && existing != nil {
				// Update collected_at only
				if err := tx.Model(&bronze.GCPComputeGlobalAddress{}).
					Where("resource_id = ?", addr.ResourceID).
					Update("collected_at", addr.CollectedAt).Error; err != nil {
					return fmt.Errorf("failed to update collected_at for global address %s: %w", addr.Name, err)
				}
				continue
			}

			// Delete old relations (manual cascade)
			if existing != nil {
				if err := s.deleteGlobalAddressRelations(tx, addr.ResourceID); err != nil {
					return fmt.Errorf("failed to delete old relations for global address %s: %w", addr.Name, err)
				}
			}

			// Upsert address
			if err := tx.Save(&addr).Error; err != nil {
				return fmt.Errorf("failed to upsert global address %s: %w", addr.Name, err)
			}

			// Create new relations
			if err := s.createGlobalAddressRelations(tx, addr.ResourceID, &addr); err != nil {
				return fmt.Errorf("failed to create relations for global address %s: %w", addr.Name, err)
			}

			// Track history
			if diff.IsNew {
				if err := s.history.CreateHistory(tx, &addr, now); err != nil {
					return fmt.Errorf("failed to create history for global address %s: %w", addr.Name, err)
				}
			} else {
				if err := s.history.UpdateHistory(tx, existing, &addr, diff, now); err != nil {
					return fmt.Errorf("failed to update history for global address %s: %w", addr.Name, err)
				}
			}
		}

		return nil
	})
}

// deleteGlobalAddressRelations deletes all related records for a global address.
func (s *Service) deleteGlobalAddressRelations(tx *gorm.DB, globalAddressResourceID string) error {
	if err := tx.Where("global_address_resource_id = ?", globalAddressResourceID).Delete(&bronze.GCPComputeGlobalAddressLabel{}).Error; err != nil {
		return err
	}
	return nil
}

// createGlobalAddressRelations creates all related records for a global address.
func (s *Service) createGlobalAddressRelations(tx *gorm.DB, globalAddressResourceID string, addr *bronze.GCPComputeGlobalAddress) error {
	for i := range addr.Labels {
		addr.Labels[i].GlobalAddressResourceID = globalAddressResourceID
	}
	if len(addr.Labels) > 0 {
		if err := tx.Create(&addr.Labels).Error; err != nil {
			return fmt.Errorf("failed to create labels: %w", err)
		}
	}
	return nil
}

// DeleteStaleGlobalAddresses removes global addresses that were not collected in the latest run.
// Also closes history records for deleted addresses.
func (s *Service) DeleteStaleGlobalAddresses(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find stale global addresses
		var staleAddresses []bronze.GCPComputeGlobalAddress
		if err := tx.Where("project_id = ? AND collected_at < ?", projectID, collectedAt).
			Find(&staleAddresses).Error; err != nil {
			return err
		}

		// Close history and delete each stale address
		for _, a := range staleAddresses {
			// Close history
			if err := s.history.CloseHistory(tx, a.ResourceID, now); err != nil {
				return fmt.Errorf("failed to close history for global address %s: %w", a.ResourceID, err)
			}

			// Delete relations
			if err := s.deleteGlobalAddressRelations(tx, a.ResourceID); err != nil {
				return fmt.Errorf("failed to delete relations for global address %s: %w", a.ResourceID, err)
			}

			// Delete address
			if err := tx.Delete(&a).Error; err != nil {
				return fmt.Errorf("failed to delete global address %s: %w", a.ResourceID, err)
			}
		}

		return nil
	})
}
