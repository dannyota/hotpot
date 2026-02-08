package vpngateway

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorygcpvpngateway"
	"hotpot/pkg/storage/ent/bronzehistorygcpvpngatewaylabel"
)

// HistoryService handles history tracking for VPN gateways.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new VPN gateway and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *VpnGatewayData, now time.Time) error {
	// Create VPN gateway history
	create := tx.BronzeHistoryGCPVPNGateway.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetDescription(data.Description).
		SetRegion(data.Region).
		SetNetwork(data.Network).
		SetSelfLink(data.SelfLink).
		SetCreationTimestamp(data.CreationTimestamp).
		SetLabelFingerprint(data.LabelFingerprint).
		SetGatewayIPVersion(data.GatewayIpVersion).
		SetStackType(data.StackType).
		SetProjectID(data.ProjectID)

	if len(data.VpnInterfacesJSON) > 0 {
		create.SetVpnInterfacesJSON(data.VpnInterfacesJSON)
	}

	gwHist, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("create vpn gateway history: %w", err)
	}

	// Create labels history
	for _, label := range data.Labels {
		_, err := tx.BronzeHistoryGCPVPNGatewayLabel.Create().
			SetVpnGatewayHistoryID(gwHist.HistoryID).
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
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPVPNGateway, new *VpnGatewayData, diff *VpnGatewayDiff, now time.Time) error {
	if !diff.IsChanged && !diff.LabelsDiff.HasChanges {
		return nil
	}

	// Get current VPN gateway history
	currentHist, err := tx.BronzeHistoryGCPVPNGateway.Query().
		Where(
			bronzehistorygcpvpngateway.ResourceID(old.ID),
			bronzehistorygcpvpngateway.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("query current history: %w", err)
	}

	// If VPN gateway-level fields changed, close old and create new VPN gateway history
	if diff.IsChanged {
		// Close old VPN gateway history
		_, err := tx.BronzeHistoryGCPVPNGateway.UpdateOneID(int(currentHist.HistoryID)).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("close old history: %w", err)
		}

		// Close all children history
		_, err = tx.BronzeHistoryGCPVPNGatewayLabel.Update().
			Where(
				bronzehistorygcpvpngatewaylabel.VpnGatewayHistoryID(currentHist.HistoryID),
				bronzehistorygcpvpngatewaylabel.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("close children history: %w", err)
		}

		// Create new history
		return h.CreateHistory(ctx, tx, new, now)
	}

	// VPN gateway unchanged, update children individually (granular tracking)
	if diff.LabelsDiff.HasChanges {
		// Close old labels history
		_, err := tx.BronzeHistoryGCPVPNGatewayLabel.Update().
			Where(
				bronzehistorygcpvpngatewaylabel.VpnGatewayHistoryID(currentHist.HistoryID),
				bronzehistorygcpvpngatewaylabel.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("close old labels history: %w", err)
		}

		// Create new labels history
		for _, label := range new.Labels {
			_, err := tx.BronzeHistoryGCPVPNGatewayLabel.Create().
				SetVpnGatewayHistoryID(currentHist.HistoryID).
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

// CloseHistory closes history records for a deleted VPN gateway.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	// Get current VPN gateway history
	currentHist, err := tx.BronzeHistoryGCPVPNGateway.Query().
		Where(
			bronzehistorygcpvpngateway.ResourceID(resourceID),
			bronzehistorygcpvpngateway.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("query current history: %w", err)
	}

	// Close VPN gateway history
	_, err = tx.BronzeHistoryGCPVPNGateway.UpdateOneID(int(currentHist.HistoryID)).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close vpn gateway history: %w", err)
	}

	// Close all children history
	_, err = tx.BronzeHistoryGCPVPNGatewayLabel.Update().
		Where(
			bronzehistorygcpvpngatewaylabel.VpnGatewayHistoryID(currentHist.HistoryID),
			bronzehistorygcpvpngatewaylabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close children history: %w", err)
	}

	return nil
}
