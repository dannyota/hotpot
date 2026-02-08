package network

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputenetwork"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputenetworkpeering"
)

// HistoryService handles history tracking for networks.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new network and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, networkData *NetworkData, now time.Time) error {
	// Create network history
	netHist, err := tx.BronzeHistoryGCPComputeNetwork.Create().
		SetResourceID(networkData.ID).
		SetValidFrom(now).
		SetCollectedAt(networkData.CollectedAt).
		SetName(networkData.Name).
		SetDescription(networkData.Description).
		SetSelfLink(networkData.SelfLink).
		SetCreationTimestamp(networkData.CreationTimestamp).
		SetAutoCreateSubnetworks(networkData.AutoCreateSubnetworks).
		SetMtu(networkData.Mtu).
		SetRoutingMode(networkData.RoutingMode).
		SetNetworkFirewallPolicyEnforcementOrder(networkData.NetworkFirewallPolicyEnforcementOrder).
		SetEnableUlaInternalIpv6(networkData.EnableUlaInternalIpv6).
		SetInternalIpv6Range(networkData.InternalIpv6Range).
		SetGatewayIpv4(networkData.GatewayIpv4).
		SetSubnetworksJSON(networkData.SubnetworksJSON).
		SetProjectID(networkData.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create network history: %w", err)
	}

	// Create children history with network_history_id
	return h.createChildrenHistory(ctx, tx, netHist.HistoryID, networkData, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeNetwork, new *NetworkData, diff *NetworkDiff, now time.Time) error {
	// Get current network history
	currentHist, err := tx.BronzeHistoryGCPComputeNetwork.Query().
		Where(
			bronzehistorygcpcomputenetwork.ResourceID(old.ID),
			bronzehistorygcpcomputenetwork.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current network history: %w", err)
	}

	// If network-level fields changed, close old and create new network history
	if diff.IsChanged {
		// Close old network history
		err = tx.BronzeHistoryGCPComputeNetwork.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current network history: %w", err)
		}

		// Create new network history
		netHist, err := tx.BronzeHistoryGCPComputeNetwork.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetName(new.Name).
			SetDescription(new.Description).
			SetSelfLink(new.SelfLink).
			SetCreationTimestamp(new.CreationTimestamp).
			SetAutoCreateSubnetworks(new.AutoCreateSubnetworks).
			SetMtu(new.Mtu).
			SetRoutingMode(new.RoutingMode).
			SetNetworkFirewallPolicyEnforcementOrder(new.NetworkFirewallPolicyEnforcementOrder).
			SetEnableUlaInternalIpv6(new.EnableUlaInternalIpv6).
			SetInternalIpv6Range(new.InternalIpv6Range).
			SetGatewayIpv4(new.GatewayIpv4).
			SetSubnetworksJSON(new.SubnetworksJSON).
			SetProjectID(new.ProjectID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new network history: %w", err)
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(ctx, tx, currentHist.HistoryID, now); err != nil {
			return fmt.Errorf("failed to close children history: %w", err)
		}
		return h.createChildrenHistory(ctx, tx, netHist.HistoryID, new, now)
	}

	// Network unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(ctx, tx, currentHist.HistoryID, new, diff, now)
}

// CloseHistory closes history records for a deleted network.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	// Get current network history
	currentHist, err := tx.BronzeHistoryGCPComputeNetwork.Query().
		Where(
			bronzehistorygcpcomputenetwork.ResourceID(resourceID),
			bronzehistorygcpcomputenetwork.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("failed to find current network history: %w", err)
	}

	// Close network history
	err = tx.BronzeHistoryGCPComputeNetwork.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close network history: %w", err)
	}

	// Close all children history
	return h.closeChildrenHistory(ctx, tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(ctx context.Context, tx *ent.Tx, networkHistoryID uint, network *NetworkData, now time.Time) error {
	// Peerings
	for _, peering := range network.Peerings {
		_, err := tx.BronzeHistoryGCPComputeNetworkPeering.Create().
			SetNetworkHistoryID(networkHistoryID).
			SetValidFrom(now).
			SetName(peering.Name).
			SetNetwork(peering.Network).
			SetState(peering.State).
			SetStateDetails(peering.StateDetails).
			SetExportCustomRoutes(peering.ExportCustomRoutes).
			SetImportCustomRoutes(peering.ImportCustomRoutes).
			SetExportSubnetRoutesWithPublicIP(peering.ExportSubnetRoutesWithPublicIp).
			SetImportSubnetRoutesWithPublicIP(peering.ImportSubnetRoutesWithPublicIp).
			SetExchangeSubnetRoutes(peering.ExchangeSubnetRoutes).
			SetStackType(peering.StackType).
			SetPeerMtu(peering.PeerMtu).
			SetAutoCreateRoutes(peering.AutoCreateRoutes).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create peering history: %w", err)
		}
	}

	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(ctx context.Context, tx *ent.Tx, networkHistoryID uint, now time.Time) error {
	// Close peerings
	_, err := tx.BronzeHistoryGCPComputeNetworkPeering.Update().
		Where(
			bronzehistorygcpcomputenetworkpeering.NetworkHistoryID(networkHistoryID),
			bronzehistorygcpcomputenetworkpeering.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close peering history: %w", err)
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(ctx context.Context, tx *ent.Tx, networkHistoryID uint, new *NetworkData, diff *NetworkDiff, now time.Time) error {
	if diff.PeeringsDiff.Changed {
		if err := h.updatePeeringsHistory(ctx, tx, networkHistoryID, new.Peerings, now); err != nil {
			return fmt.Errorf("failed to update peerings history: %w", err)
		}
	}

	return nil
}

func (h *HistoryService) updatePeeringsHistory(ctx context.Context, tx *ent.Tx, networkHistoryID uint, peerings []PeeringData, now time.Time) error {
	// Close old peering history
	_, err := tx.BronzeHistoryGCPComputeNetworkPeering.Update().
		Where(
			bronzehistorygcpcomputenetworkpeering.NetworkHistoryID(networkHistoryID),
			bronzehistorygcpcomputenetworkpeering.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close peering history: %w", err)
	}

	// Create new peering history
	for _, peering := range peerings {
		_, err := tx.BronzeHistoryGCPComputeNetworkPeering.Create().
			SetNetworkHistoryID(networkHistoryID).
			SetValidFrom(now).
			SetName(peering.Name).
			SetNetwork(peering.Network).
			SetState(peering.State).
			SetStateDetails(peering.StateDetails).
			SetExportCustomRoutes(peering.ExportCustomRoutes).
			SetImportCustomRoutes(peering.ImportCustomRoutes).
			SetExportSubnetRoutesWithPublicIP(peering.ExportSubnetRoutesWithPublicIp).
			SetImportSubnetRoutesWithPublicIP(peering.ImportSubnetRoutesWithPublicIp).
			SetExchangeSubnetRoutes(peering.ExchangeSubnetRoutes).
			SetStackType(peering.StackType).
			SetPeerMtu(peering.PeerMtu).
			SetAutoCreateRoutes(peering.AutoCreateRoutes).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create peering history: %w", err)
		}
	}

	return nil
}
