package address

import (
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
	"hotpot/pkg/base/models/bronze_history"
)

// HistoryService handles history tracking for addresses.
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService creates a new history service.
func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

// CreateHistory creates history records for a new address and all children.
func (h *HistoryService) CreateHistory(tx *gorm.DB, addr *bronze.GCPComputeAddress, now time.Time) error {
	// Create address history
	addrHist := toAddressHistory(addr, now)
	if err := tx.Create(&addrHist).Error; err != nil {
		return err
	}

	// Create children history with address_history_id
	return h.createChildrenHistory(tx, addrHist.HistoryID, addr, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(tx *gorm.DB, old, new *bronze.GCPComputeAddress, diff *AddressDiff, now time.Time) error {
	// Get current address history
	var currentHist bronze_history.GCPComputeAddress
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", old.ResourceID).First(&currentHist).Error; err != nil {
		return err
	}

	// If address-level fields changed, close old and create new address history
	if diff.IsChanged {
		// Close old address history
		if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
			return err
		}

		// Create new address history
		addrHist := toAddressHistory(new, now)
		if err := tx.Create(&addrHist).Error; err != nil {
			return err
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(tx, currentHist.HistoryID, now); err != nil {
			return err
		}
		return h.createChildrenHistory(tx, addrHist.HistoryID, new, now)
	}

	// Address unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(tx, currentHist.HistoryID, new, diff, now)
}

// CloseHistory closes history records for a deleted address.
func (h *HistoryService) CloseHistory(tx *gorm.DB, resourceID string, now time.Time) error {
	// Get current address history
	var currentHist bronze_history.GCPComputeAddress
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", resourceID).First(&currentHist).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // No history to close
		}
		return err
	}

	// Close address history
	if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
		return err
	}

	// Close all children history
	return h.closeChildrenHistory(tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(tx *gorm.DB, addressHistoryID uint, addr *bronze.GCPComputeAddress, now time.Time) error {
	for _, label := range addr.Labels {
		labelHist := toLabelHistory(&label, addressHistoryID, now)
		if err := tx.Create(&labelHist).Error; err != nil {
			return err
		}
	}
	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(tx *gorm.DB, addressHistoryID uint, now time.Time) error {
	if err := tx.Table("bronze_history.gcp_compute_address_labels").
		Where("address_history_id = ? AND valid_to IS NULL", addressHistoryID).
		Update("valid_to", now).Error; err != nil {
		return err
	}
	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(tx *gorm.DB, addressHistoryID uint, new *bronze.GCPComputeAddress, diff *AddressDiff, now time.Time) error {
	if diff.LabelsDiff.Changed {
		if err := h.updateLabelsHistory(tx, addressHistoryID, new.Labels, now); err != nil {
			return err
		}
	}
	return nil
}

func (h *HistoryService) updateLabelsHistory(tx *gorm.DB, addressHistoryID uint, labels []bronze.GCPComputeAddressLabel, now time.Time) error {
	if err := tx.Table("bronze_history.gcp_compute_address_labels").
		Where("address_history_id = ? AND valid_to IS NULL", addressHistoryID).
		Update("valid_to", now).Error; err != nil {
		return err
	}

	for _, label := range labels {
		labelHist := toLabelHistory(&label, addressHistoryID, now)
		if err := tx.Create(&labelHist).Error; err != nil {
			return err
		}
	}
	return nil
}

// Conversion functions: bronze -> bronze_history

func toAddressHistory(addr *bronze.GCPComputeAddress, now time.Time) bronze_history.GCPComputeAddress {
	return bronze_history.GCPComputeAddress{
		ResourceID:        addr.ResourceID,
		ValidFrom:         now,
		ValidTo:           nil,
		Name:              addr.Name,
		Description:       addr.Description,
		Address:           addr.Address,
		AddressType:       addr.AddressType,
		IpVersion:         addr.IpVersion,
		Ipv6EndpointType:  addr.Ipv6EndpointType,
		IpCollection:      addr.IpCollection,
		Region:            addr.Region,
		Status:            addr.Status,
		Purpose:           addr.Purpose,
		Network:           addr.Network,
		Subnetwork:        addr.Subnetwork,
		NetworkTier:       addr.NetworkTier,
		PrefixLength:      addr.PrefixLength,
		SelfLink:          addr.SelfLink,
		CreationTimestamp: addr.CreationTimestamp,
		LabelFingerprint:  addr.LabelFingerprint,
		UsersJSON:         addr.UsersJSON,
		ProjectID:         addr.ProjectID,
		CollectedAt:       addr.CollectedAt,
	}
}

func toLabelHistory(label *bronze.GCPComputeAddressLabel, addressHistoryID uint, now time.Time) bronze_history.GCPComputeAddressLabel {
	return bronze_history.GCPComputeAddressLabel{
		AddressHistoryID: addressHistoryID,
		ValidFrom:        now,
		ValidTo:          nil,
		Key:              label.Key,
		Value:            label.Value,
	}
}
