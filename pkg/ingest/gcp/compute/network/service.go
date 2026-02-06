package network

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
)

// Service handles GCP Compute network ingestion.
type Service struct {
	client  *Client
	db      *gorm.DB
	history *HistoryService
}

// NewService creates a new network ingestion service.
func NewService(client *Client, db *gorm.DB) *Service {
	return &Service{
		client:  client,
		db:      db,
		history: NewHistoryService(db),
	}
}

// IngestParams contains parameters for network ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of network ingestion.
type IngestResult struct {
	ProjectID      string
	NetworkCount   int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches networks from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch networks from GCP
	networks, err := s.client.ListNetworks(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	// Convert to bronze models
	bronzeNetworks := make([]bronze.GCPComputeNetwork, 0, len(networks))
	for _, n := range networks {
		bn, err := ConvertNetwork(n, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert network: %w", err)
		}
		bronzeNetworks = append(bronzeNetworks, bn)
	}

	// Save to database
	if err := s.saveNetworks(ctx, bronzeNetworks); err != nil {
		return nil, fmt.Errorf("failed to save networks: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		NetworkCount:   len(bronzeNetworks),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveNetworks saves networks to the database with history tracking.
func (s *Service) saveNetworks(ctx context.Context, networks []bronze.GCPComputeNetwork) error {
	if len(networks) == 0 {
		return nil
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, network := range networks {
			// Load existing network with all relations
			var existing *bronze.GCPComputeNetwork
			var old bronze.GCPComputeNetwork
			err := tx.Preload("Peerings").
				Where("resource_id = ?", network.ResourceID).
				First(&old).Error
			if err == nil {
				existing = &old
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to load existing network %s: %w", network.Name, err)
			}

			// Compute diff
			diff := DiffNetwork(existing, &network)

			// Skip if no changes
			if !diff.HasAnyChange() && existing != nil {
				// Update collected_at only
				if err := tx.Model(&bronze.GCPComputeNetwork{}).
					Where("resource_id = ?", network.ResourceID).
					Update("collected_at", network.CollectedAt).Error; err != nil {
					return fmt.Errorf("failed to update collected_at for network %s: %w", network.Name, err)
				}
				continue
			}

			// Delete old relations (manual cascade)
			if existing != nil {
				if err := s.deleteNetworkRelations(tx, network.ResourceID); err != nil {
					return fmt.Errorf("failed to delete old relations for network %s: %w", network.Name, err)
				}
			}

			// Upsert network
			if err := tx.Save(&network).Error; err != nil {
				return fmt.Errorf("failed to upsert network %s: %w", network.Name, err)
			}

			// Create new relations
			if err := s.createNetworkRelations(tx, network.ResourceID, &network); err != nil {
				return fmt.Errorf("failed to create relations for network %s: %w", network.Name, err)
			}

			// Track history
			if diff.IsNew {
				if err := s.history.CreateHistory(tx, &network, now); err != nil {
					return fmt.Errorf("failed to create history for network %s: %w", network.Name, err)
				}
			} else {
				if err := s.history.UpdateHistory(tx, existing, &network, diff, now); err != nil {
					return fmt.Errorf("failed to update history for network %s: %w", network.Name, err)
				}
			}
		}

		return nil
	})
}

// deleteNetworkRelations deletes all related records for a network.
func (s *Service) deleteNetworkRelations(tx *gorm.DB, networkResourceID string) error {
	// Delete peerings
	if err := tx.Where("network_resource_id = ?", networkResourceID).Delete(&bronze.GCPComputeNetworkPeering{}).Error; err != nil {
		return err
	}

	return nil
}

// createNetworkRelations creates all related records for a network.
func (s *Service) createNetworkRelations(tx *gorm.DB, networkResourceID string, network *bronze.GCPComputeNetwork) error {
	// Create peerings
	for i := range network.Peerings {
		network.Peerings[i].NetworkResourceID = networkResourceID
	}
	if len(network.Peerings) > 0 {
		if err := tx.Create(&network.Peerings).Error; err != nil {
			return fmt.Errorf("failed to create peerings: %w", err)
		}
	}

	return nil
}

// DeleteStaleNetworks removes networks that were not collected in the latest run.
// Also closes history records for deleted networks.
func (s *Service) DeleteStaleNetworks(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find stale networks
		var staleNetworks []bronze.GCPComputeNetwork
		if err := tx.Where("project_id = ? AND collected_at < ?", projectID, collectedAt).
			Find(&staleNetworks).Error; err != nil {
			return err
		}

		// Close history and delete each stale network
		for _, n := range staleNetworks {
			// Close history
			if err := s.history.CloseHistory(tx, n.ResourceID, now); err != nil {
				return fmt.Errorf("failed to close history for network %s: %w", n.ResourceID, err)
			}

			// Delete relations
			if err := s.deleteNetworkRelations(tx, n.ResourceID); err != nil {
				return fmt.Errorf("failed to delete relations for network %s: %w", n.ResourceID, err)
			}

			// Delete network
			if err := tx.Delete(&n).Error; err != nil {
				return fmt.Errorf("failed to delete network %s: %w", n.ResourceID, err)
			}
		}

		return nil
	})
}
