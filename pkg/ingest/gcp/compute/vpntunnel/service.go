package vpntunnel

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
)

// Service handles GCP Compute VPN tunnel ingestion.
type Service struct {
	client  *Client
	db      *gorm.DB
	history *HistoryService
}

// NewService creates a new VPN tunnel ingestion service.
func NewService(client *Client, db *gorm.DB) *Service {
	return &Service{
		client:  client,
		db:      db,
		history: NewHistoryService(db),
	}
}

// IngestParams contains parameters for VPN tunnel ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of VPN tunnel ingestion.
type IngestResult struct {
	ProjectID      string
	VpnTunnelCount int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches VPN tunnels from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch VPN tunnels from GCP
	tunnels, err := s.client.ListVpnTunnels(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list vpn tunnels: %w", err)
	}

	// Convert to bronze models
	bronzeTunnels := make([]bronze.GCPComputeVpnTunnel, 0, len(tunnels))
	for _, t := range tunnels {
		bt, err := ConvertVpnTunnel(t, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert vpn tunnel: %w", err)
		}
		bronzeTunnels = append(bronzeTunnels, bt)
	}

	// Save to database
	if err := s.saveVpnTunnels(ctx, bronzeTunnels); err != nil {
		return nil, fmt.Errorf("failed to save vpn tunnels: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		VpnTunnelCount: len(bronzeTunnels),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveVpnTunnels saves VPN tunnels to the database with history tracking.
func (s *Service) saveVpnTunnels(ctx context.Context, tunnels []bronze.GCPComputeVpnTunnel) error {
	if len(tunnels) == 0 {
		return nil
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, tunnel := range tunnels {
			// Load existing VPN tunnel with all relations
			var existing *bronze.GCPComputeVpnTunnel
			var old bronze.GCPComputeVpnTunnel
			err := tx.Preload("Labels").
				Where("resource_id = ?", tunnel.ResourceID).
				First(&old).Error
			if err == nil {
				existing = &old
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to load existing vpn tunnel %s: %w", tunnel.Name, err)
			}

			// Compute diff
			diff := DiffVpnTunnel(existing, &tunnel)

			// Skip if no changes
			if !diff.HasAnyChange() && existing != nil {
				// Update collected_at only
				if err := tx.Model(&bronze.GCPComputeVpnTunnel{}).
					Where("resource_id = ?", tunnel.ResourceID).
					Update("collected_at", tunnel.CollectedAt).Error; err != nil {
					return fmt.Errorf("failed to update collected_at for vpn tunnel %s: %w", tunnel.Name, err)
				}
				continue
			}

			// Delete old relations (manual cascade)
			if existing != nil {
				if err := s.deleteVpnTunnelRelations(tx, tunnel.ResourceID); err != nil {
					return fmt.Errorf("failed to delete old relations for vpn tunnel %s: %w", tunnel.Name, err)
				}
			}

			// Upsert VPN tunnel
			if err := tx.Save(&tunnel).Error; err != nil {
				return fmt.Errorf("failed to upsert vpn tunnel %s: %w", tunnel.Name, err)
			}

			// Create new relations
			if err := s.createVpnTunnelRelations(tx, tunnel.ResourceID, &tunnel); err != nil {
				return fmt.Errorf("failed to create relations for vpn tunnel %s: %w", tunnel.Name, err)
			}

			// Track history
			if diff.IsNew {
				if err := s.history.CreateHistory(tx, &tunnel, now); err != nil {
					return fmt.Errorf("failed to create history for vpn tunnel %s: %w", tunnel.Name, err)
				}
			} else {
				if err := s.history.UpdateHistory(tx, existing, &tunnel, diff, now); err != nil {
					return fmt.Errorf("failed to update history for vpn tunnel %s: %w", tunnel.Name, err)
				}
			}
		}

		return nil
	})
}

// deleteVpnTunnelRelations deletes all related records for a VPN tunnel.
func (s *Service) deleteVpnTunnelRelations(tx *gorm.DB, vpnTunnelResourceID string) error {
	// Delete labels
	if err := tx.Where("vpn_tunnel_resource_id = ?", vpnTunnelResourceID).Delete(&bronze.GCPComputeVpnTunnelLabel{}).Error; err != nil {
		return err
	}

	return nil
}

// createVpnTunnelRelations creates all related records for a VPN tunnel.
func (s *Service) createVpnTunnelRelations(tx *gorm.DB, vpnTunnelResourceID string, tunnel *bronze.GCPComputeVpnTunnel) error {
	// Create labels
	for i := range tunnel.Labels {
		tunnel.Labels[i].VpnTunnelResourceID = vpnTunnelResourceID
	}
	if len(tunnel.Labels) > 0 {
		if err := tx.Create(&tunnel.Labels).Error; err != nil {
			return fmt.Errorf("failed to create labels: %w", err)
		}
	}

	return nil
}

// DeleteStaleVpnTunnels removes VPN tunnels that were not collected in the latest run.
// Also closes history records for deleted VPN tunnels.
func (s *Service) DeleteStaleVpnTunnels(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find stale VPN tunnels
		var staleTunnels []bronze.GCPComputeVpnTunnel
		if err := tx.Where("project_id = ? AND collected_at < ?", projectID, collectedAt).
			Find(&staleTunnels).Error; err != nil {
			return err
		}

		// Close history and delete each stale VPN tunnel
		for _, tunnel := range staleTunnels {
			// Close history
			if err := s.history.CloseHistory(tx, tunnel.ResourceID, now); err != nil {
				return fmt.Errorf("failed to close history for vpn tunnel %s: %w", tunnel.ResourceID, err)
			}

			// Delete relations
			if err := s.deleteVpnTunnelRelations(tx, tunnel.ResourceID); err != nil {
				return fmt.Errorf("failed to delete relations for vpn tunnel %s: %w", tunnel.ResourceID, err)
			}

			// Delete VPN tunnel
			if err := tx.Delete(&tunnel).Error; err != nil {
				return fmt.Errorf("failed to delete vpn tunnel %s: %w", tunnel.ResourceID, err)
			}
		}

		return nil
	})
}
