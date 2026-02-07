package targetvpngateway

import (
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
	"hotpot/pkg/base/models/bronze_history"
)

// HistoryService handles history tracking for Classic VPN gateways.
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService creates a new history service.
func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

// CreateHistory creates history records for a new target VPN gateway and all children.
func (h *HistoryService) CreateHistory(tx *gorm.DB, gw *bronze.GCPComputeTargetVpnGateway, now time.Time) error {
	// Create target VPN gateway history
	gwHist := toTargetVpnGatewayHistory(gw, now)
	if err := tx.Create(&gwHist).Error; err != nil {
		return err
	}

	// Create children history with target_vpn_gateway_history_id
	return h.createChildrenHistory(tx, gwHist.HistoryID, gw, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(tx *gorm.DB, old, new *bronze.GCPComputeTargetVpnGateway, diff *TargetVpnGatewayDiff, now time.Time) error {
	// Get current target VPN gateway history
	var currentHist bronze_history.GCPComputeTargetVpnGateway
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", old.ResourceID).First(&currentHist).Error; err != nil {
		return err
	}

	// If target VPN gateway-level fields changed, close old and create new history
	if diff.IsChanged {
		// Close old target VPN gateway history
		if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
			return err
		}

		// Create new target VPN gateway history
		gwHist := toTargetVpnGatewayHistory(new, now)
		if err := tx.Create(&gwHist).Error; err != nil {
			return err
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(tx, currentHist.HistoryID, now); err != nil {
			return err
		}
		return h.createChildrenHistory(tx, gwHist.HistoryID, new, now)
	}

	// Target VPN gateway unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(tx, currentHist.HistoryID, new, diff, now)
}

// CloseHistory closes history records for a deleted target VPN gateway.
func (h *HistoryService) CloseHistory(tx *gorm.DB, resourceID string, now time.Time) error {
	// Get current target VPN gateway history
	var currentHist bronze_history.GCPComputeTargetVpnGateway
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", resourceID).First(&currentHist).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // No history to close
		}
		return err
	}

	// Close target VPN gateway history
	if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
		return err
	}

	// Close all children history
	return h.closeChildrenHistory(tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(tx *gorm.DB, targetVpnGatewayHistoryID uint, gw *bronze.GCPComputeTargetVpnGateway, now time.Time) error {
	// Labels
	for _, label := range gw.Labels {
		labelHist := toLabelHistory(&label, targetVpnGatewayHistoryID, now)
		if err := tx.Create(&labelHist).Error; err != nil {
			return err
		}
	}

	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(tx *gorm.DB, targetVpnGatewayHistoryID uint, now time.Time) error {
	// Close labels
	if err := tx.Table("bronze_history.gcp_compute_target_vpn_gateway_labels").
		Where("target_vpn_gateway_history_id = ? AND valid_to IS NULL", targetVpnGatewayHistoryID).
		Update("valid_to", now).Error; err != nil {
		return err
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(tx *gorm.DB, targetVpnGatewayHistoryID uint, new *bronze.GCPComputeTargetVpnGateway, diff *TargetVpnGatewayDiff, now time.Time) error {
	if diff.LabelsDiff.Changed {
		if err := h.updateLabelsHistory(tx, targetVpnGatewayHistoryID, new.Labels, now); err != nil {
			return err
		}
	}

	return nil
}

func (h *HistoryService) updateLabelsHistory(tx *gorm.DB, targetVpnGatewayHistoryID uint, labels []bronze.GCPComputeTargetVpnGatewayLabel, now time.Time) error {
	tx.Table("bronze_history.gcp_compute_target_vpn_gateway_labels").
		Where("target_vpn_gateway_history_id = ? AND valid_to IS NULL", targetVpnGatewayHistoryID).
		Update("valid_to", now)

	for _, label := range labels {
		labelHist := toLabelHistory(&label, targetVpnGatewayHistoryID, now)
		if err := tx.Create(&labelHist).Error; err != nil {
			return err
		}
	}
	return nil
}

// Conversion functions: bronze -> bronze_history

func toTargetVpnGatewayHistory(gw *bronze.GCPComputeTargetVpnGateway, now time.Time) bronze_history.GCPComputeTargetVpnGateway {
	return bronze_history.GCPComputeTargetVpnGateway{
		ResourceID:          gw.ResourceID,
		ValidFrom:           now,
		ValidTo:             nil,
		Name:                gw.Name,
		Description:         gw.Description,
		Status:              gw.Status,
		Region:              gw.Region,
		Network:             gw.Network,
		SelfLink:            gw.SelfLink,
		CreationTimestamp:   gw.CreationTimestamp,
		LabelFingerprint:    gw.LabelFingerprint,
		ForwardingRulesJSON: gw.ForwardingRulesJSON,
		TunnelsJSON:         gw.TunnelsJSON,
		ProjectID:           gw.ProjectID,
		CollectedAt:         gw.CollectedAt,
	}
}

func toLabelHistory(label *bronze.GCPComputeTargetVpnGatewayLabel, targetVpnGatewayHistoryID uint, now time.Time) bronze_history.GCPComputeTargetVpnGatewayLabel {
	return bronze_history.GCPComputeTargetVpnGatewayLabel{
		TargetVpnGatewayHistoryID: targetVpnGatewayHistoryID,
		ValidFrom:                 now,
		ValidTo:                   nil,
		Key:                       label.Key,
		Value:                     label.Value,
	}
}
