package vpntunnel

import (
	"context"
	"fmt"
	"time"

	entvpn "github.com/dannyota/hotpot/pkg/storage/ent/gcp/vpn"
	"github.com/dannyota/hotpot/pkg/storage/ent/gcp/vpn/bronzehistorygcpvpntunnel"
	"github.com/dannyota/hotpot/pkg/storage/ent/gcp/vpn/bronzehistorygcpvpntunnellabel"
)

// HistoryService handles history tracking for VPN tunnels.
type HistoryService struct {
	entClient *entvpn.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entvpn.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new VPN tunnel and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entvpn.Tx, data *VpnTunnelData, now time.Time) error {
	// Create VPN tunnel history
	create := tx.BronzeHistoryGCPVPNTunnel.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
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
		SetProjectID(data.ProjectID)

	if len(data.LocalTrafficSelectorJSON) > 0 {
		create.SetLocalTrafficSelectorJSON(data.LocalTrafficSelectorJSON)
	}
	if len(data.RemoteTrafficSelectorJSON) > 0 {
		create.SetRemoteTrafficSelectorJSON(data.RemoteTrafficSelectorJSON)
	}

	tunnelHist, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("create vpn tunnel history: %w", err)
	}

	// Create labels history
	for _, label := range data.Labels {
		_, err := tx.BronzeHistoryGCPVPNTunnelLabel.Create().
			SetVpnTunnelHistoryID(tunnelHist.ID).
			SetValidFrom(now).
			SetKey(label.Key).
			SetValue(label.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("create label history: %w", err)
		}
	}

	return nil
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entvpn.Tx, old *entvpn.BronzeGCPVPNTunnel, new *VpnTunnelData, diff *VpnTunnelDiff, now time.Time) error {
	if !diff.IsChanged && !diff.LabelsDiff.HasChanges {
		return nil
	}

	// Get current VPN tunnel history
	currentHist, err := tx.BronzeHistoryGCPVPNTunnel.Query().
		Where(
			bronzehistorygcpvpntunnel.ResourceID(old.ID),
			bronzehistorygcpvpntunnel.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("query current history: %w", err)
	}

	// If VPN tunnel-level fields changed, close old and create new VPN tunnel history
	if diff.IsChanged {
		// Close old VPN tunnel history
		_, err := tx.BronzeHistoryGCPVPNTunnel.UpdateOneID(currentHist.ID).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("close old history: %w", err)
		}

		// Close all children history
		_, err = tx.BronzeHistoryGCPVPNTunnelLabel.Update().
			Where(
				bronzehistorygcpvpntunnellabel.VpnTunnelHistoryID(currentHist.ID),
				bronzehistorygcpvpntunnellabel.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("close children history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPVPNTunnel.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetDescription(new.Description).
			SetStatus(new.Status).
			SetDetailedStatus(new.DetailedStatus).
			SetRegion(new.Region).
			SetSelfLink(new.SelfLink).
			SetCreationTimestamp(new.CreationTimestamp).
			SetLabelFingerprint(new.LabelFingerprint).
			SetIkeVersion(new.IkeVersion).
			SetPeerIP(new.PeerIp).
			SetPeerExternalGateway(new.PeerExternalGateway).
			SetPeerExternalGatewayInterface(new.PeerExternalGatewayInterface).
			SetPeerGcpGateway(new.PeerGcpGateway).
			SetRouter(new.Router).
			SetSharedSecretHash(new.SharedSecretHash).
			SetVpnGateway(new.VpnGateway).
			SetTargetVpnGateway(new.TargetVpnGateway).
			SetVpnGatewayInterface(new.VpnGatewayInterface).
			SetProjectID(new.ProjectID)

		if len(new.LocalTrafficSelectorJSON) > 0 {
			create.SetLocalTrafficSelectorJSON(new.LocalTrafficSelectorJSON)
		}
		if len(new.RemoteTrafficSelectorJSON) > 0 {
			create.SetRemoteTrafficSelectorJSON(new.RemoteTrafficSelectorJSON)
		}

		tunnelHist, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("create vpn tunnel history: %w", err)
		}

		// Create labels history
		for _, label := range new.Labels {
			_, err := tx.BronzeHistoryGCPVPNTunnelLabel.Create().
				SetVpnTunnelHistoryID(tunnelHist.ID).
				SetValidFrom(now).
				SetKey(label.Key).
				SetValue(label.Value).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("create label history: %w", err)
			}
		}

		return nil
	}

	// VPN tunnel unchanged, update children individually (granular tracking)
	if diff.LabelsDiff.HasChanges {
		// Close old labels history
		_, err := tx.BronzeHistoryGCPVPNTunnelLabel.Update().
			Where(
				bronzehistorygcpvpntunnellabel.VpnTunnelHistoryID(currentHist.ID),
				bronzehistorygcpvpntunnellabel.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("close old labels history: %w", err)
		}

		// Create new labels history
		for _, label := range new.Labels {
			_, err := tx.BronzeHistoryGCPVPNTunnelLabel.Create().
				SetVpnTunnelHistoryID(currentHist.ID).
				SetValidFrom(now).
				SetKey(label.Key).
				SetValue(label.Value).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("create label history: %w", err)
			}
		}
	}

	return nil
}

// CloseHistory closes history records for a deleted VPN tunnel.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entvpn.Tx, resourceID string, now time.Time) error {
	// Get current VPN tunnel history
	currentHist, err := tx.BronzeHistoryGCPVPNTunnel.Query().
		Where(
			bronzehistorygcpvpntunnel.ResourceID(resourceID),
			bronzehistorygcpvpntunnel.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entvpn.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("query current history: %w", err)
	}

	// Close VPN tunnel history
	_, err = tx.BronzeHistoryGCPVPNTunnel.UpdateOneID(currentHist.ID).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close vpn tunnel history: %w", err)
	}

	// Close all children history
	_, err = tx.BronzeHistoryGCPVPNTunnelLabel.Update().
		Where(
			bronzehistorygcpvpntunnellabel.VpnTunnelHistoryID(currentHist.ID),
			bronzehistorygcpvpntunnellabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close children history: %w", err)
	}

	return nil
}
