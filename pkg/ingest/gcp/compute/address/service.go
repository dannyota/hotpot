package address

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
)

// Service handles GCP Compute address ingestion.
type Service struct {
	client  *Client
	db      *gorm.DB
	history *HistoryService
}

// NewService creates a new address ingestion service.
func NewService(client *Client, db *gorm.DB) *Service {
	return &Service{
		client:  client,
		db:      db,
		history: NewHistoryService(db),
	}
}

// IngestParams contains parameters for address ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of address ingestion.
type IngestResult struct {
	ProjectID      string
	AddressCount   int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches addresses from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch addresses from GCP
	addresses, err := s.client.ListAddresses(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list addresses: %w", err)
	}

	// Convert to bronze models
	bronzeAddresses := make([]bronze.GCPComputeAddress, 0, len(addresses))
	for _, a := range addresses {
		addr, err := ConvertAddress(a, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert address: %w", err)
		}
		bronzeAddresses = append(bronzeAddresses, addr)
	}

	// Save to database
	if err := s.saveAddresses(ctx, bronzeAddresses); err != nil {
		return nil, fmt.Errorf("failed to save addresses: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		AddressCount:   len(bronzeAddresses),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveAddresses saves addresses to the database with history tracking.
func (s *Service) saveAddresses(ctx context.Context, addresses []bronze.GCPComputeAddress) error {
	if len(addresses) == 0 {
		return nil
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, addr := range addresses {
			// Load existing address with all relations
			var existing *bronze.GCPComputeAddress
			var old bronze.GCPComputeAddress
			err := tx.Preload("Labels").
				Where("resource_id = ?", addr.ResourceID).
				First(&old).Error
			if err == nil {
				existing = &old
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to load existing address %s: %w", addr.Name, err)
			}

			// Compute diff
			diff := DiffAddress(existing, &addr)

			// Skip if no changes
			if !diff.HasAnyChange() && existing != nil {
				// Update collected_at only
				if err := tx.Model(&bronze.GCPComputeAddress{}).
					Where("resource_id = ?", addr.ResourceID).
					Update("collected_at", addr.CollectedAt).Error; err != nil {
					return fmt.Errorf("failed to update collected_at for address %s: %w", addr.Name, err)
				}
				continue
			}

			// Delete old relations (manual cascade)
			if existing != nil {
				if err := s.deleteAddressRelations(tx, addr.ResourceID); err != nil {
					return fmt.Errorf("failed to delete old relations for address %s: %w", addr.Name, err)
				}
			}

			// Upsert address
			if err := tx.Save(&addr).Error; err != nil {
				return fmt.Errorf("failed to upsert address %s: %w", addr.Name, err)
			}

			// Create new relations
			if err := s.createAddressRelations(tx, addr.ResourceID, &addr); err != nil {
				return fmt.Errorf("failed to create relations for address %s: %w", addr.Name, err)
			}

			// Track history
			if diff.IsNew {
				if err := s.history.CreateHistory(tx, &addr, now); err != nil {
					return fmt.Errorf("failed to create history for address %s: %w", addr.Name, err)
				}
			} else {
				if err := s.history.UpdateHistory(tx, existing, &addr, diff, now); err != nil {
					return fmt.Errorf("failed to update history for address %s: %w", addr.Name, err)
				}
			}
		}

		return nil
	})
}

// deleteAddressRelations deletes all related records for an address.
func (s *Service) deleteAddressRelations(tx *gorm.DB, addressResourceID string) error {
	if err := tx.Where("address_resource_id = ?", addressResourceID).Delete(&bronze.GCPComputeAddressLabel{}).Error; err != nil {
		return err
	}
	return nil
}

// createAddressRelations creates all related records for an address.
func (s *Service) createAddressRelations(tx *gorm.DB, addressResourceID string, addr *bronze.GCPComputeAddress) error {
	for i := range addr.Labels {
		addr.Labels[i].AddressResourceID = addressResourceID
	}
	if len(addr.Labels) > 0 {
		if err := tx.Create(&addr.Labels).Error; err != nil {
			return fmt.Errorf("failed to create labels: %w", err)
		}
	}
	return nil
}

// DeleteStaleAddresses removes addresses that were not collected in the latest run.
// Also closes history records for deleted addresses.
func (s *Service) DeleteStaleAddresses(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find stale addresses
		var staleAddresses []bronze.GCPComputeAddress
		if err := tx.Where("project_id = ? AND collected_at < ?", projectID, collectedAt).
			Find(&staleAddresses).Error; err != nil {
			return err
		}

		// Close history and delete each stale address
		for _, a := range staleAddresses {
			// Close history
			if err := s.history.CloseHistory(tx, a.ResourceID, now); err != nil {
				return fmt.Errorf("failed to close history for address %s: %w", a.ResourceID, err)
			}

			// Delete relations
			if err := s.deleteAddressRelations(tx, a.ResourceID); err != nil {
				return fmt.Errorf("failed to delete relations for address %s: %w", a.ResourceID, err)
			}

			// Delete address
			if err := tx.Delete(&a).Error; err != nil {
				return fmt.Errorf("failed to delete address %s: %w", a.ResourceID, err)
			}
		}

		return nil
	})
}
