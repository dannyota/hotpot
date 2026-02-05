package network

import (
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
	"hotpot/pkg/base/models/bronze_history"
)

// HistoryService handles history tracking for networks.
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService creates a new history service.
func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

// CreateHistory creates history records for a new network and all children.
func (h *HistoryService) CreateHistory(tx *gorm.DB, network *bronze.GCPComputeNetwork, now time.Time) error {
	// Create network history
	netHist := toNetworkHistory(network, now)
	if err := tx.Create(&netHist).Error; err != nil {
		return err
	}

	// Create children history with network_history_id
	return h.createChildrenHistory(tx, netHist.HistoryID, network, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(tx *gorm.DB, old, new *bronze.GCPComputeNetwork, diff *NetworkDiff, now time.Time) error {
	// Get current network history
	var currentHist bronze_history.GCPComputeNetwork
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", old.ResourceID).First(&currentHist).Error; err != nil {
		return err
	}

	// If network-level fields changed, close old and create new network history
	if diff.IsChanged {
		// Close old network history
		if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
			return err
		}

		// Create new network history
		netHist := toNetworkHistory(new, now)
		if err := tx.Create(&netHist).Error; err != nil {
			return err
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(tx, currentHist.HistoryID, now); err != nil {
			return err
		}
		return h.createChildrenHistory(tx, netHist.HistoryID, new, now)
	}

	// Network unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(tx, currentHist.HistoryID, new, diff, now)
}

// CloseHistory closes history records for a deleted network.
func (h *HistoryService) CloseHistory(tx *gorm.DB, resourceID string, now time.Time) error {
	// Get current network history
	var currentHist bronze_history.GCPComputeNetwork
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", resourceID).First(&currentHist).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // No history to close
		}
		return err
	}

	// Close network history
	if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
		return err
	}

	// Close all children history
	return h.closeChildrenHistory(tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(tx *gorm.DB, networkHistoryID uint, network *bronze.GCPComputeNetwork, now time.Time) error {
	// Peerings
	for _, peering := range network.Peerings {
		peerHist := toPeeringHistory(&peering, networkHistoryID, now)
		if err := tx.Create(&peerHist).Error; err != nil {
			return err
		}
	}

	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(tx *gorm.DB, networkHistoryID uint, now time.Time) error {
	// Close peerings
	if err := tx.Table("bronze_history.gcp_compute_network_peerings").
		Where("network_history_id = ? AND valid_to IS NULL", networkHistoryID).
		Update("valid_to", now).Error; err != nil {
		return err
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(tx *gorm.DB, networkHistoryID uint, new *bronze.GCPComputeNetwork, diff *NetworkDiff, now time.Time) error {
	if diff.PeeringsDiff.Changed {
		if err := h.updatePeeringsHistory(tx, networkHistoryID, new.Peerings, now); err != nil {
			return err
		}
	}

	return nil
}

func (h *HistoryService) updatePeeringsHistory(tx *gorm.DB, networkHistoryID uint, peerings []bronze.GCPComputeNetworkPeering, now time.Time) error {
	tx.Table("bronze_history.gcp_compute_network_peerings").
		Where("network_history_id = ? AND valid_to IS NULL", networkHistoryID).
		Update("valid_to", now)

	for _, peering := range peerings {
		peerHist := toPeeringHistory(&peering, networkHistoryID, now)
		if err := tx.Create(&peerHist).Error; err != nil {
			return err
		}
	}
	return nil
}

// Conversion functions: bronze -> bronze_history

func toNetworkHistory(network *bronze.GCPComputeNetwork, now time.Time) bronze_history.GCPComputeNetwork {
	return bronze_history.GCPComputeNetwork{
		ResourceID:                            network.ResourceID,
		ValidFrom:                             now,
		ValidTo:                               nil,
		Name:                                  network.Name,
		Description:                           network.Description,
		SelfLink:                              network.SelfLink,
		CreationTimestamp:                     network.CreationTimestamp,
		AutoCreateSubnetworks:                 network.AutoCreateSubnetworks,
		Mtu:                                   network.Mtu,
		RoutingMode:                           network.RoutingMode,
		NetworkFirewallPolicyEnforcementOrder: network.NetworkFirewallPolicyEnforcementOrder,
		EnableUlaInternalIpv6:                 network.EnableUlaInternalIpv6,
		InternalIpv6Range:                     network.InternalIpv6Range,
		GatewayIpv4:                           network.GatewayIpv4,
		SubnetworksJSON:                       network.SubnetworksJSON,
		ProjectID:                             network.ProjectID,
		CollectedAt:                           network.CollectedAt,
	}
}

func toPeeringHistory(peering *bronze.GCPComputeNetworkPeering, networkHistoryID uint, now time.Time) bronze_history.GCPComputeNetworkPeering {
	return bronze_history.GCPComputeNetworkPeering{
		NetworkHistoryID:               networkHistoryID,
		ValidFrom:                      now,
		ValidTo:                        nil,
		Name:                           peering.Name,
		Network:                        peering.Network,
		State:                          peering.State,
		StateDetails:                   peering.StateDetails,
		ExportCustomRoutes:             peering.ExportCustomRoutes,
		ImportCustomRoutes:             peering.ImportCustomRoutes,
		ExportSubnetRoutesWithPublicIp: peering.ExportSubnetRoutesWithPublicIp,
		ImportSubnetRoutesWithPublicIp: peering.ImportSubnetRoutesWithPublicIp,
		ExchangeSubnetRoutes:           peering.ExchangeSubnetRoutes,
		StackType:                      peering.StackType,
		PeerMtu:                        peering.PeerMtu,
		AutoCreateRoutes:               peering.AutoCreateRoutes,
	}
}
