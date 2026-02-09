package vpntunnel

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpvpntunnel"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpvpntunnellabel"
)

// Service handles GCP Compute VPN tunnel ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new VPN tunnel ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
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

	// Convert to data structs
	vpnTunnelDataList := make([]*VpnTunnelData, 0, len(tunnels))
	for _, t := range tunnels {
		data, err := ConvertVpnTunnel(t, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert vpn tunnel: %w", err)
		}
		vpnTunnelDataList = append(vpnTunnelDataList, data)
	}

	// Save to database
	if err := s.saveVpnTunnels(ctx, vpnTunnelDataList); err != nil {
		return nil, fmt.Errorf("failed to save vpn tunnels: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		VpnTunnelCount: len(vpnTunnelDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveVpnTunnels saves VPN tunnels to the database with history tracking.
func (s *Service) saveVpnTunnels(ctx context.Context, vpnTunnels []*VpnTunnelData) error {
	if len(vpnTunnels) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, data := range vpnTunnels {
		// Load existing VPN tunnel with all relations
		existing, err := tx.BronzeGCPVPNTunnel.Query().
			Where(bronzegcpvpntunnel.ID(data.ID)).
			WithLabels().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("query existing vpn tunnel %s: %w", data.Name, err)
		}

		// Compute diff
		diff := DiffVpnTunnelData(existing, data)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			_, err := tx.BronzeGCPVPNTunnel.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for vpn tunnel %s: %w", data.Name, err)
			}
			continue
		}

		// Upsert VPN tunnel
		var savedTunnel *ent.BronzeGCPVPNTunnel
		if existing == nil {
			// Create new VPN tunnel
			create := tx.BronzeGCPVPNTunnel.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetStatus(data.Status).
				SetDetailedStatus(data.DetailedStatus).
				SetRegion(data.Region).
				SetSelfLink(data.SelfLink).
				SetCreationTimestamp(data.CreationTimestamp).
				SetLabelFingerprint(data.LabelFingerprint).
				SetIkeVersion(data.IkeVersion).
				SetPeerIP(data.PeerIp).
				SetPeerExternalGateway(data.PeerExternalGateway).
				SetPeerExternalGatewayInterface(data.PeerExternalGatewayInterface).
				SetPeerGcpGateway(data.PeerGcpGateway).
				SetRouter(data.Router).
				SetSharedSecretHash(data.SharedSecretHash).
				SetVpnGateway(data.VpnGateway).
				SetTargetVpnGateway(data.TargetVpnGateway).
				SetVpnGatewayInterface(data.VpnGatewayInterface).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if len(data.LocalTrafficSelectorJSON) > 0 {
				create.SetLocalTrafficSelectorJSON(data.LocalTrafficSelectorJSON)
			}
			if len(data.RemoteTrafficSelectorJSON) > 0 {
				create.SetRemoteTrafficSelectorJSON(data.RemoteTrafficSelectorJSON)
			}

			savedTunnel, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create vpn tunnel %s: %w", data.Name, err)
			}
		} else {
			// Update existing VPN tunnel
			update := tx.BronzeGCPVPNTunnel.UpdateOneID(data.ID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetStatus(data.Status).
				SetDetailedStatus(data.DetailedStatus).
				SetRegion(data.Region).
				SetSelfLink(data.SelfLink).
				SetCreationTimestamp(data.CreationTimestamp).
				SetLabelFingerprint(data.LabelFingerprint).
				SetIkeVersion(data.IkeVersion).
				SetPeerIP(data.PeerIp).
				SetPeerExternalGateway(data.PeerExternalGateway).
				SetPeerExternalGatewayInterface(data.PeerExternalGatewayInterface).
				SetPeerGcpGateway(data.PeerGcpGateway).
				SetRouter(data.Router).
				SetSharedSecretHash(data.SharedSecretHash).
				SetVpnGateway(data.VpnGateway).
				SetTargetVpnGateway(data.TargetVpnGateway).
				SetVpnGatewayInterface(data.VpnGatewayInterface).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt)

			if len(data.LocalTrafficSelectorJSON) > 0 {
				update.SetLocalTrafficSelectorJSON(data.LocalTrafficSelectorJSON)
			} else {
				update.ClearLocalTrafficSelectorJSON()
			}
			if len(data.RemoteTrafficSelectorJSON) > 0 {
				update.SetRemoteTrafficSelectorJSON(data.RemoteTrafficSelectorJSON)
			} else {
				update.ClearRemoteTrafficSelectorJSON()
			}

			savedTunnel, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update vpn tunnel %s: %w", data.Name, err)
			}

			// Delete old labels (cascade handled by edges)
			_, err = tx.BronzeGCPVPNTunnelLabel.Delete().
				Where(bronzegcpvpntunnellabel.HasVpnTunnelWith(bronzegcpvpntunnel.ID(data.ID))).
				Exec(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("delete old labels for vpn tunnel %s: %w", data.Name, err)
			}
		}

		// Create labels
		for _, labelData := range data.Labels {
			_, err := tx.BronzeGCPVPNTunnelLabel.Create().
				SetKey(labelData.Key).
				SetValue(labelData.Value).
				SetVpnTunnel(savedTunnel).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create label for vpn tunnel %s: %w", data.Name, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for vpn tunnel %s: %w", data.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, data, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for vpn tunnel %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleVpnTunnels removes VPN tunnels that were not collected in the latest run.
// Also closes history records for deleted VPN tunnels.
func (s *Service) DeleteStaleVpnTunnels(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	// Find stale VPN tunnels
	staleVpnTunnels, err := tx.BronzeGCPVPNTunnel.Query().
		Where(
			bronzegcpvpntunnel.ProjectID(projectID),
			bronzegcpvpntunnel.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale vpn tunnels: %w", err)
	}

	// Close history and delete each stale VPN tunnel
	for _, tunnel := range staleVpnTunnels {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, tunnel.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for vpn tunnel %s: %w", tunnel.ID, err)
		}

		// Delete VPN tunnel (labels cascade via edge)
		if err := tx.BronzeGCPVPNTunnel.DeleteOneID(tunnel.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete vpn tunnel %s: %w", tunnel.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
