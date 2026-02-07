package vpntunnel

import (
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
	"hotpot/pkg/base/models/bronze_history"
)

// HistoryService handles history tracking for VPN tunnels.
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService creates a new history service.
func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

// CreateHistory creates history records for a new VPN tunnel and all children.
func (h *HistoryService) CreateHistory(tx *gorm.DB, tunnel *bronze.GCPComputeVpnTunnel, now time.Time) error {
	// Create VPN tunnel history
	tunnelHist := toVpnTunnelHistory(tunnel, now)
	if err := tx.Create(&tunnelHist).Error; err != nil {
		return err
	}

	// Create children history with vpn_tunnel_history_id
	return h.createChildrenHistory(tx, tunnelHist.HistoryID, tunnel, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(tx *gorm.DB, old, new *bronze.GCPComputeVpnTunnel, diff *VpnTunnelDiff, now time.Time) error {
	// Get current VPN tunnel history
	var currentHist bronze_history.GCPComputeVpnTunnel
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", old.ResourceID).First(&currentHist).Error; err != nil {
		return err
	}

	// If VPN tunnel-level fields changed, close old and create new history
	if diff.IsChanged {
		// Close old VPN tunnel history
		if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
			return err
		}

		// Create new VPN tunnel history
		tunnelHist := toVpnTunnelHistory(new, now)
		if err := tx.Create(&tunnelHist).Error; err != nil {
			return err
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(tx, currentHist.HistoryID, now); err != nil {
			return err
		}
		return h.createChildrenHistory(tx, tunnelHist.HistoryID, new, now)
	}

	// VPN tunnel unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(tx, currentHist.HistoryID, new, diff, now)
}

// CloseHistory closes history records for a deleted VPN tunnel.
func (h *HistoryService) CloseHistory(tx *gorm.DB, resourceID string, now time.Time) error {
	// Get current VPN tunnel history
	var currentHist bronze_history.GCPComputeVpnTunnel
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", resourceID).First(&currentHist).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // No history to close
		}
		return err
	}

	// Close VPN tunnel history
	if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
		return err
	}

	// Close all children history
	return h.closeChildrenHistory(tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(tx *gorm.DB, vpnTunnelHistoryID uint, tunnel *bronze.GCPComputeVpnTunnel, now time.Time) error {
	// Labels
	for _, label := range tunnel.Labels {
		labelHist := toLabelHistory(&label, vpnTunnelHistoryID, now)
		if err := tx.Create(&labelHist).Error; err != nil {
			return err
		}
	}

	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(tx *gorm.DB, vpnTunnelHistoryID uint, now time.Time) error {
	// Close labels
	if err := tx.Table("bronze_history.gcp_compute_vpn_tunnel_labels").
		Where("vpn_tunnel_history_id = ? AND valid_to IS NULL", vpnTunnelHistoryID).
		Update("valid_to", now).Error; err != nil {
		return err
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(tx *gorm.DB, vpnTunnelHistoryID uint, new *bronze.GCPComputeVpnTunnel, diff *VpnTunnelDiff, now time.Time) error {
	if diff.LabelsDiff.Changed {
		if err := h.updateLabelsHistory(tx, vpnTunnelHistoryID, new.Labels, now); err != nil {
			return err
		}
	}

	return nil
}

func (h *HistoryService) updateLabelsHistory(tx *gorm.DB, vpnTunnelHistoryID uint, labels []bronze.GCPComputeVpnTunnelLabel, now time.Time) error {
	tx.Table("bronze_history.gcp_compute_vpn_tunnel_labels").
		Where("vpn_tunnel_history_id = ? AND valid_to IS NULL", vpnTunnelHistoryID).
		Update("valid_to", now)

	for _, label := range labels {
		labelHist := toLabelHistory(&label, vpnTunnelHistoryID, now)
		if err := tx.Create(&labelHist).Error; err != nil {
			return err
		}
	}
	return nil
}

// Conversion functions: bronze -> bronze_history

func toVpnTunnelHistory(tunnel *bronze.GCPComputeVpnTunnel, now time.Time) bronze_history.GCPComputeVpnTunnel {
	return bronze_history.GCPComputeVpnTunnel{
		ResourceID:                   tunnel.ResourceID,
		ValidFrom:                    now,
		ValidTo:                      nil,
		Name:                         tunnel.Name,
		Description:                  tunnel.Description,
		Status:                       tunnel.Status,
		DetailedStatus:               tunnel.DetailedStatus,
		Region:                       tunnel.Region,
		SelfLink:                     tunnel.SelfLink,
		CreationTimestamp:            tunnel.CreationTimestamp,
		LabelFingerprint:             tunnel.LabelFingerprint,
		IkeVersion:                   tunnel.IkeVersion,
		PeerIp:                       tunnel.PeerIp,
		PeerExternalGateway:          tunnel.PeerExternalGateway,
		PeerExternalGatewayInterface: tunnel.PeerExternalGatewayInterface,
		PeerGcpGateway:               tunnel.PeerGcpGateway,
		Router:                       tunnel.Router,
		SharedSecretHash:             tunnel.SharedSecretHash,
		VpnGateway:                   tunnel.VpnGateway,
		TargetVpnGateway:             tunnel.TargetVpnGateway,
		VpnGatewayInterface:          tunnel.VpnGatewayInterface,
		LocalTrafficSelectorJSON:     tunnel.LocalTrafficSelectorJSON,
		RemoteTrafficSelectorJSON:    tunnel.RemoteTrafficSelectorJSON,
		ProjectID:                    tunnel.ProjectID,
		CollectedAt:                  tunnel.CollectedAt,
	}
}

func toLabelHistory(label *bronze.GCPComputeVpnTunnelLabel, vpnTunnelHistoryID uint, now time.Time) bronze_history.GCPComputeVpnTunnelLabel {
	return bronze_history.GCPComputeVpnTunnelLabel{
		VpnTunnelHistoryID: vpnTunnelHistoryID,
		ValidFrom:          now,
		ValidTo:            nil,
		Key:                label.Key,
		Value:              label.Value,
	}
}
