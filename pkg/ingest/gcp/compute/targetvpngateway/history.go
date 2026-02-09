package targetvpngateway

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpvpntargetgateway"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpvpntargetgatewaylabel"
)

// HistoryService handles history tracking for Classic VPN gateways.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new target VPN gateway and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *TargetVpnGatewayData, now time.Time) error {
	// Create target VPN gateway history
	create := tx.BronzeHistoryGCPVPNTargetGateway.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetDescription(data.Description).
		SetStatus(data.Status).
		SetRegion(data.Region).
		SetNetwork(data.Network).
		SetSelfLink(data.SelfLink).
		SetCreationTimestamp(data.CreationTimestamp).
		SetLabelFingerprint(data.LabelFingerprint).
		SetProjectID(data.ProjectID)

	if len(data.ForwardingRulesJSON) > 0 {
		create.SetForwardingRulesJSON(data.ForwardingRulesJSON)
	}

	if len(data.TunnelsJSON) > 0 {
		create.SetTunnelsJSON(data.TunnelsJSON)
	}

	gwHist, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("create target vpn gateway history: %w", err)
	}

	// Create labels history
	for _, label := range data.Labels {
		_, err := tx.BronzeHistoryGCPVPNTargetGatewayLabel.Create().
			SetTargetVpnGatewayHistoryID(gwHist.HistoryID).
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
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPVPNTargetGateway, new *TargetVpnGatewayData, diff *TargetVpnGatewayDiff, now time.Time) error {
	if !diff.IsChanged && !diff.LabelsDiff.HasChanges {
		return nil
	}

	// Get current target VPN gateway history
	currentHist, err := tx.BronzeHistoryGCPVPNTargetGateway.Query().
		Where(
			bronzehistorygcpvpntargetgateway.ResourceID(old.ID),
			bronzehistorygcpvpntargetgateway.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("query current history: %w", err)
	}

	// If target VPN gateway-level fields changed, close old and create new target VPN gateway history
	if diff.IsChanged {
		// Close old target VPN gateway history
		_, err := tx.BronzeHistoryGCPVPNTargetGateway.UpdateOneID(int(currentHist.HistoryID)).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("close old history: %w", err)
		}

		// Close all children history
		_, err = tx.BronzeHistoryGCPVPNTargetGatewayLabel.Update().
			Where(
				bronzehistorygcpvpntargetgatewaylabel.TargetVpnGatewayHistoryID(currentHist.HistoryID),
				bronzehistorygcpvpntargetgatewaylabel.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("close children history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPVPNTargetGateway.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetDescription(new.Description).
			SetStatus(new.Status).
			SetRegion(new.Region).
			SetNetwork(new.Network).
			SetSelfLink(new.SelfLink).
			SetCreationTimestamp(new.CreationTimestamp).
			SetLabelFingerprint(new.LabelFingerprint).
			SetProjectID(new.ProjectID)

		if len(new.ForwardingRulesJSON) > 0 {
			create.SetForwardingRulesJSON(new.ForwardingRulesJSON)
		}

		if len(new.TunnelsJSON) > 0 {
			create.SetTunnelsJSON(new.TunnelsJSON)
		}

		gwHist, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("create target vpn gateway history: %w", err)
		}

		// Create labels history
		for _, label := range new.Labels {
			_, err := tx.BronzeHistoryGCPVPNTargetGatewayLabel.Create().
				SetTargetVpnGatewayHistoryID(gwHist.HistoryID).
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

	// Target VPN gateway unchanged, update children individually (granular tracking)
	if diff.LabelsDiff.HasChanges {
		// Close old labels history
		_, err := tx.BronzeHistoryGCPVPNTargetGatewayLabel.Update().
			Where(
				bronzehistorygcpvpntargetgatewaylabel.TargetVpnGatewayHistoryID(currentHist.HistoryID),
				bronzehistorygcpvpntargetgatewaylabel.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("close old labels history: %w", err)
		}

		// Create new labels history
		for _, label := range new.Labels {
			_, err := tx.BronzeHistoryGCPVPNTargetGatewayLabel.Create().
				SetTargetVpnGatewayHistoryID(currentHist.HistoryID).
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

// CloseHistory closes history records for a deleted target VPN gateway.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	// Get current target VPN gateway history
	currentHist, err := tx.BronzeHistoryGCPVPNTargetGateway.Query().
		Where(
			bronzehistorygcpvpntargetgateway.ResourceID(resourceID),
			bronzehistorygcpvpntargetgateway.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("query current history: %w", err)
	}

	// Close target VPN gateway history
	_, err = tx.BronzeHistoryGCPVPNTargetGateway.UpdateOneID(int(currentHist.HistoryID)).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close target vpn gateway history: %w", err)
	}

	// Close all children history
	_, err = tx.BronzeHistoryGCPVPNTargetGatewayLabel.Update().
		Where(
			bronzehistorygcpvpntargetgatewaylabel.TargetVpnGatewayHistoryID(currentHist.HistoryID),
			bronzehistorygcpvpntargetgatewaylabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close children history: %w", err)
	}

	return nil
}
