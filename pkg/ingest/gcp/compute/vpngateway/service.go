package vpngateway

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
)

// Service handles GCP Compute VPN gateway ingestion.
type Service struct {
	client  *Client
	db      *gorm.DB
	history *HistoryService
}

// NewService creates a new VPN gateway ingestion service.
func NewService(client *Client, db *gorm.DB) *Service {
	return &Service{
		client:  client,
		db:      db,
		history: NewHistoryService(db),
	}
}

// IngestParams contains parameters for VPN gateway ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of VPN gateway ingestion.
type IngestResult struct {
	ProjectID       string
	VpnGatewayCount int
	CollectedAt     time.Time
	DurationMillis  int64
}

// Ingest fetches VPN gateways from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch VPN gateways from GCP
	vpnGateways, err := s.client.ListVpnGateways(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list vpn gateways: %w", err)
	}

	// Convert to bronze models
	bronzeVpnGateways := make([]bronze.GCPComputeVpnGateway, 0, len(vpnGateways))
	for _, gw := range vpnGateways {
		bg, err := ConvertVpnGateway(gw, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert vpn gateway: %w", err)
		}
		bronzeVpnGateways = append(bronzeVpnGateways, bg)
	}

	// Save to database
	if err := s.saveVpnGateways(ctx, bronzeVpnGateways); err != nil {
		return nil, fmt.Errorf("failed to save vpn gateways: %w", err)
	}

	return &IngestResult{
		ProjectID:       params.ProjectID,
		VpnGatewayCount: len(bronzeVpnGateways),
		CollectedAt:     collectedAt,
		DurationMillis:  time.Since(startTime).Milliseconds(),
	}, nil
}

// saveVpnGateways saves VPN gateways to the database with history tracking.
func (s *Service) saveVpnGateways(ctx context.Context, vpnGateways []bronze.GCPComputeVpnGateway) error {
	if len(vpnGateways) == 0 {
		return nil
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, gw := range vpnGateways {
			// Load existing VPN gateway with all relations
			var existing *bronze.GCPComputeVpnGateway
			var old bronze.GCPComputeVpnGateway
			err := tx.Preload("Labels").
				Where("resource_id = ?", gw.ResourceID).
				First(&old).Error
			if err == nil {
				existing = &old
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to load existing vpn gateway %s: %w", gw.Name, err)
			}

			// Compute diff
			diff := DiffVpnGateway(existing, &gw)

			// Skip if no changes
			if !diff.HasAnyChange() && existing != nil {
				// Update collected_at only
				if err := tx.Model(&bronze.GCPComputeVpnGateway{}).
					Where("resource_id = ?", gw.ResourceID).
					Update("collected_at", gw.CollectedAt).Error; err != nil {
					return fmt.Errorf("failed to update collected_at for vpn gateway %s: %w", gw.Name, err)
				}
				continue
			}

			// Delete old relations (manual cascade)
			if existing != nil {
				if err := s.deleteVpnGatewayRelations(tx, gw.ResourceID); err != nil {
					return fmt.Errorf("failed to delete old relations for vpn gateway %s: %w", gw.Name, err)
				}
			}

			// Upsert VPN gateway
			if err := tx.Save(&gw).Error; err != nil {
				return fmt.Errorf("failed to upsert vpn gateway %s: %w", gw.Name, err)
			}

			// Create new relations
			if err := s.createVpnGatewayRelations(tx, gw.ResourceID, &gw); err != nil {
				return fmt.Errorf("failed to create relations for vpn gateway %s: %w", gw.Name, err)
			}

			// Track history
			if diff.IsNew {
				if err := s.history.CreateHistory(tx, &gw, now); err != nil {
					return fmt.Errorf("failed to create history for vpn gateway %s: %w", gw.Name, err)
				}
			} else {
				if err := s.history.UpdateHistory(tx, existing, &gw, diff, now); err != nil {
					return fmt.Errorf("failed to update history for vpn gateway %s: %w", gw.Name, err)
				}
			}
		}

		return nil
	})
}

// deleteVpnGatewayRelations deletes all related records for a VPN gateway.
func (s *Service) deleteVpnGatewayRelations(tx *gorm.DB, vpnGatewayResourceID string) error {
	// Delete labels
	if err := tx.Where("vpn_gateway_resource_id = ?", vpnGatewayResourceID).Delete(&bronze.GCPComputeVpnGatewayLabel{}).Error; err != nil {
		return err
	}

	return nil
}

// createVpnGatewayRelations creates all related records for a VPN gateway.
func (s *Service) createVpnGatewayRelations(tx *gorm.DB, vpnGatewayResourceID string, gw *bronze.GCPComputeVpnGateway) error {
	// Create labels
	for i := range gw.Labels {
		gw.Labels[i].VpnGatewayResourceID = vpnGatewayResourceID
	}
	if len(gw.Labels) > 0 {
		if err := tx.Create(&gw.Labels).Error; err != nil {
			return fmt.Errorf("failed to create labels: %w", err)
		}
	}

	return nil
}

// DeleteStaleVpnGateways removes VPN gateways that were not collected in the latest run.
// Also closes history records for deleted VPN gateways.
func (s *Service) DeleteStaleVpnGateways(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find stale VPN gateways
		var staleVpnGateways []bronze.GCPComputeVpnGateway
		if err := tx.Where("project_id = ? AND collected_at < ?", projectID, collectedAt).
			Find(&staleVpnGateways).Error; err != nil {
			return err
		}

		// Close history and delete each stale VPN gateway
		for _, gw := range staleVpnGateways {
			// Close history
			if err := s.history.CloseHistory(tx, gw.ResourceID, now); err != nil {
				return fmt.Errorf("failed to close history for vpn gateway %s: %w", gw.ResourceID, err)
			}

			// Delete relations
			if err := s.deleteVpnGatewayRelations(tx, gw.ResourceID); err != nil {
				return fmt.Errorf("failed to delete relations for vpn gateway %s: %w", gw.ResourceID, err)
			}

			// Delete VPN gateway
			if err := tx.Delete(&gw).Error; err != nil {
				return fmt.Errorf("failed to delete vpn gateway %s: %w", gw.ResourceID, err)
			}
		}

		return nil
	})
}
