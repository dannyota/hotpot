package vpngateway

import (
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
	"hotpot/pkg/base/models/bronze_history"
)

// HistoryService handles history tracking for VPN gateways.
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService creates a new history service.
func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

// CreateHistory creates history records for a new VPN gateway and all children.
func (h *HistoryService) CreateHistory(tx *gorm.DB, gw *bronze.GCPComputeVpnGateway, now time.Time) error {
	// Create VPN gateway history
	gwHist := toVpnGatewayHistory(gw, now)
	if err := tx.Create(&gwHist).Error; err != nil {
		return err
	}

	// Create children history with vpn_gateway_history_id
	return h.createChildrenHistory(tx, gwHist.HistoryID, gw, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(tx *gorm.DB, old, new *bronze.GCPComputeVpnGateway, diff *VpnGatewayDiff, now time.Time) error {
	// Get current VPN gateway history
	var currentHist bronze_history.GCPComputeVpnGateway
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", old.ResourceID).First(&currentHist).Error; err != nil {
		return err
	}

	// If VPN gateway-level fields changed, close old and create new VPN gateway history
	if diff.IsChanged {
		// Close old VPN gateway history
		if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
			return err
		}

		// Create new VPN gateway history
		gwHist := toVpnGatewayHistory(new, now)
		if err := tx.Create(&gwHist).Error; err != nil {
			return err
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(tx, currentHist.HistoryID, now); err != nil {
			return err
		}
		return h.createChildrenHistory(tx, gwHist.HistoryID, new, now)
	}

	// VPN gateway unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(tx, currentHist.HistoryID, new, diff, now)
}

// CloseHistory closes history records for a deleted VPN gateway.
func (h *HistoryService) CloseHistory(tx *gorm.DB, resourceID string, now time.Time) error {
	// Get current VPN gateway history
	var currentHist bronze_history.GCPComputeVpnGateway
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", resourceID).First(&currentHist).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // No history to close
		}
		return err
	}

	// Close VPN gateway history
	if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
		return err
	}

	// Close all children history
	return h.closeChildrenHistory(tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(tx *gorm.DB, vpnGatewayHistoryID uint, gw *bronze.GCPComputeVpnGateway, now time.Time) error {
	// Labels
	for _, label := range gw.Labels {
		labelHist := toLabelHistory(&label, vpnGatewayHistoryID, now)
		if err := tx.Create(&labelHist).Error; err != nil {
			return err
		}
	}

	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(tx *gorm.DB, vpnGatewayHistoryID uint, now time.Time) error {
	// Close labels
	if err := tx.Table("bronze_history.gcp_compute_vpn_gateway_labels").
		Where("vpn_gateway_history_id = ? AND valid_to IS NULL", vpnGatewayHistoryID).
		Update("valid_to", now).Error; err != nil {
		return err
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(tx *gorm.DB, vpnGatewayHistoryID uint, new *bronze.GCPComputeVpnGateway, diff *VpnGatewayDiff, now time.Time) error {
	if diff.LabelsDiff.Changed {
		if err := h.updateLabelsHistory(tx, vpnGatewayHistoryID, new.Labels, now); err != nil {
			return err
		}
	}

	return nil
}

func (h *HistoryService) updateLabelsHistory(tx *gorm.DB, vpnGatewayHistoryID uint, labels []bronze.GCPComputeVpnGatewayLabel, now time.Time) error {
	tx.Table("bronze_history.gcp_compute_vpn_gateway_labels").
		Where("vpn_gateway_history_id = ? AND valid_to IS NULL", vpnGatewayHistoryID).
		Update("valid_to", now)

	for _, label := range labels {
		labelHist := toLabelHistory(&label, vpnGatewayHistoryID, now)
		if err := tx.Create(&labelHist).Error; err != nil {
			return err
		}
	}
	return nil
}

// Conversion functions: bronze -> bronze_history

func toVpnGatewayHistory(gw *bronze.GCPComputeVpnGateway, now time.Time) bronze_history.GCPComputeVpnGateway {
	return bronze_history.GCPComputeVpnGateway{
		ResourceID:        gw.ResourceID,
		ValidFrom:         now,
		ValidTo:           nil,
		Name:              gw.Name,
		Description:       gw.Description,
		Region:            gw.Region,
		Network:           gw.Network,
		SelfLink:          gw.SelfLink,
		CreationTimestamp: gw.CreationTimestamp,
		LabelFingerprint:  gw.LabelFingerprint,
		GatewayIpVersion:  gw.GatewayIpVersion,
		StackType:         gw.StackType,
		VpnInterfacesJSON: gw.VpnInterfacesJSON,
		ProjectID:         gw.ProjectID,
		CollectedAt:       gw.CollectedAt,
	}
}

func toLabelHistory(label *bronze.GCPComputeVpnGatewayLabel, vpnGatewayHistoryID uint, now time.Time) bronze_history.GCPComputeVpnGatewayLabel {
	return bronze_history.GCPComputeVpnGatewayLabel{
		VpnGatewayHistoryID: vpnGatewayHistoryID,
		ValidFrom:           now,
		ValidTo:             nil,
		Key:                 label.Key,
		Value:               label.Value,
	}
}
