package targetvpngateway

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
)

// Service handles GCP Compute Classic VPN gateway ingestion.
type Service struct {
	client  *Client
	db      *gorm.DB
	history *HistoryService
}

// NewService creates a new target VPN gateway ingestion service.
func NewService(client *Client, db *gorm.DB) *Service {
	return &Service{
		client:  client,
		db:      db,
		history: NewHistoryService(db),
	}
}

// IngestParams contains parameters for target VPN gateway ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of target VPN gateway ingestion.
type IngestResult struct {
	ProjectID             string
	TargetVpnGatewayCount int
	CollectedAt           time.Time
	DurationMillis        int64
}

// Ingest fetches Classic VPN gateways from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch target VPN gateways from GCP
	targetVpnGateways, err := s.client.ListTargetVpnGateways(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list target vpn gateways: %w", err)
	}

	// Convert to bronze models
	bronzeGateways := make([]bronze.GCPComputeTargetVpnGateway, 0, len(targetVpnGateways))
	for _, gw := range targetVpnGateways {
		bg, err := ConvertTargetVpnGateway(gw, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert target vpn gateway: %w", err)
		}
		bronzeGateways = append(bronzeGateways, bg)
	}

	// Save to database
	if err := s.saveTargetVpnGateways(ctx, bronzeGateways); err != nil {
		return nil, fmt.Errorf("failed to save target vpn gateways: %w", err)
	}

	return &IngestResult{
		ProjectID:             params.ProjectID,
		TargetVpnGatewayCount: len(bronzeGateways),
		CollectedAt:           collectedAt,
		DurationMillis:        time.Since(startTime).Milliseconds(),
	}, nil
}

// saveTargetVpnGateways saves target VPN gateways to the database with history tracking.
func (s *Service) saveTargetVpnGateways(ctx context.Context, gateways []bronze.GCPComputeTargetVpnGateway) error {
	if len(gateways) == 0 {
		return nil
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, gw := range gateways {
			// Load existing target VPN gateway with all relations
			var existing *bronze.GCPComputeTargetVpnGateway
			var old bronze.GCPComputeTargetVpnGateway
			err := tx.Preload("Labels").
				Where("resource_id = ?", gw.ResourceID).
				First(&old).Error
			if err == nil {
				existing = &old
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to load existing target vpn gateway %s: %w", gw.Name, err)
			}

			// Compute diff
			diff := DiffTargetVpnGateway(existing, &gw)

			// Skip if no changes
			if !diff.HasAnyChange() && existing != nil {
				// Update collected_at only
				if err := tx.Model(&bronze.GCPComputeTargetVpnGateway{}).
					Where("resource_id = ?", gw.ResourceID).
					Update("collected_at", gw.CollectedAt).Error; err != nil {
					return fmt.Errorf("failed to update collected_at for target vpn gateway %s: %w", gw.Name, err)
				}
				continue
			}

			// Delete old relations (manual cascade)
			if existing != nil {
				if err := s.deleteTargetVpnGatewayRelations(tx, gw.ResourceID); err != nil {
					return fmt.Errorf("failed to delete old relations for target vpn gateway %s: %w", gw.Name, err)
				}
			}

			// Upsert target VPN gateway
			if err := tx.Save(&gw).Error; err != nil {
				return fmt.Errorf("failed to upsert target vpn gateway %s: %w", gw.Name, err)
			}

			// Create new relations
			if err := s.createTargetVpnGatewayRelations(tx, gw.ResourceID, &gw); err != nil {
				return fmt.Errorf("failed to create relations for target vpn gateway %s: %w", gw.Name, err)
			}

			// Track history
			if diff.IsNew {
				if err := s.history.CreateHistory(tx, &gw, now); err != nil {
					return fmt.Errorf("failed to create history for target vpn gateway %s: %w", gw.Name, err)
				}
			} else {
				if err := s.history.UpdateHistory(tx, existing, &gw, diff, now); err != nil {
					return fmt.Errorf("failed to update history for target vpn gateway %s: %w", gw.Name, err)
				}
			}
		}

		return nil
	})
}

// deleteTargetVpnGatewayRelations deletes all related records for a target VPN gateway.
func (s *Service) deleteTargetVpnGatewayRelations(tx *gorm.DB, targetVpnGatewayResourceID string) error {
	// Delete labels
	if err := tx.Where("target_vpn_gateway_resource_id = ?", targetVpnGatewayResourceID).Delete(&bronze.GCPComputeTargetVpnGatewayLabel{}).Error; err != nil {
		return err
	}

	return nil
}

// createTargetVpnGatewayRelations creates all related records for a target VPN gateway.
func (s *Service) createTargetVpnGatewayRelations(tx *gorm.DB, targetVpnGatewayResourceID string, gw *bronze.GCPComputeTargetVpnGateway) error {
	// Create labels
	for i := range gw.Labels {
		gw.Labels[i].TargetVpnGatewayResourceID = targetVpnGatewayResourceID
	}
	if len(gw.Labels) > 0 {
		if err := tx.Create(&gw.Labels).Error; err != nil {
			return fmt.Errorf("failed to create labels: %w", err)
		}
	}

	return nil
}

// DeleteStaleTargetVpnGateways removes target VPN gateways that were not collected in the latest run.
// Also closes history records for deleted target VPN gateways.
func (s *Service) DeleteStaleTargetVpnGateways(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find stale target VPN gateways
		var staleGateways []bronze.GCPComputeTargetVpnGateway
		if err := tx.Where("project_id = ? AND collected_at < ?", projectID, collectedAt).
			Find(&staleGateways).Error; err != nil {
			return err
		}

		// Close history and delete each stale target VPN gateway
		for _, gw := range staleGateways {
			// Close history
			if err := s.history.CloseHistory(tx, gw.ResourceID, now); err != nil {
				return fmt.Errorf("failed to close history for target vpn gateway %s: %w", gw.ResourceID, err)
			}

			// Delete relations
			if err := s.deleteTargetVpnGatewayRelations(tx, gw.ResourceID); err != nil {
				return fmt.Errorf("failed to delete relations for target vpn gateway %s: %w", gw.ResourceID, err)
			}

			// Delete target VPN gateway
			if err := tx.Delete(&gw).Error; err != nil {
				return fmt.Errorf("failed to delete target vpn gateway %s: %w", gw.ResourceID, err)
			}
		}

		return nil
	})
}
