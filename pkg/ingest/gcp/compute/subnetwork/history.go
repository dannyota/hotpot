package subnetwork

import (
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
	"hotpot/pkg/base/models/bronze_history"
)

// HistoryService handles history tracking for subnetworks.
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService creates a new history service.
func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

// CreateHistory creates history records for a new subnetwork and all children.
func (h *HistoryService) CreateHistory(tx *gorm.DB, subnet *bronze.GCPComputeSubnetwork, now time.Time) error {
	// Create subnetwork history
	subnetHist := toSubnetworkHistory(subnet, now)
	if err := tx.Create(&subnetHist).Error; err != nil {
		return err
	}

	// Create children history with subnetwork_history_id
	return h.createChildrenHistory(tx, subnetHist.HistoryID, subnet, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(tx *gorm.DB, old, new *bronze.GCPComputeSubnetwork, diff *SubnetworkDiff, now time.Time) error {
	// Get current subnetwork history
	var currentHist bronze_history.GCPComputeSubnetwork
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", old.ResourceID).First(&currentHist).Error; err != nil {
		return err
	}

	// If subnetwork-level fields changed, close old and create new subnetwork history
	if diff.IsChanged {
		// Close old subnetwork history
		if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
			return err
		}

		// Create new subnetwork history
		subnetHist := toSubnetworkHistory(new, now)
		if err := tx.Create(&subnetHist).Error; err != nil {
			return err
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(tx, currentHist.HistoryID, now); err != nil {
			return err
		}
		return h.createChildrenHistory(tx, subnetHist.HistoryID, new, now)
	}

	// Subnetwork unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(tx, currentHist.HistoryID, new, diff, now)
}

// CloseHistory closes history records for a deleted subnetwork.
func (h *HistoryService) CloseHistory(tx *gorm.DB, resourceID string, now time.Time) error {
	// Get current subnetwork history
	var currentHist bronze_history.GCPComputeSubnetwork
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", resourceID).First(&currentHist).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // No history to close
		}
		return err
	}

	// Close subnetwork history
	if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
		return err
	}

	// Close all children history
	return h.closeChildrenHistory(tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(tx *gorm.DB, subnetworkHistoryID uint, subnet *bronze.GCPComputeSubnetwork, now time.Time) error {
	// Secondary ranges
	for _, sr := range subnet.SecondaryIpRanges {
		srHist := toSecondaryRangeHistory(&sr, subnetworkHistoryID, now)
		if err := tx.Create(&srHist).Error; err != nil {
			return err
		}
	}

	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(tx *gorm.DB, subnetworkHistoryID uint, now time.Time) error {
	// Close secondary ranges
	if err := tx.Table("bronze_history.gcp_compute_subnetwork_secondary_ranges").
		Where("subnetwork_history_id = ? AND valid_to IS NULL", subnetworkHistoryID).
		Update("valid_to", now).Error; err != nil {
		return err
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(tx *gorm.DB, subnetworkHistoryID uint, new *bronze.GCPComputeSubnetwork, diff *SubnetworkDiff, now time.Time) error {
	if diff.SecondaryRangesDiff.Changed {
		if err := h.updateSecondaryRangesHistory(tx, subnetworkHistoryID, new.SecondaryIpRanges, now); err != nil {
			return err
		}
	}

	return nil
}

func (h *HistoryService) updateSecondaryRangesHistory(tx *gorm.DB, subnetworkHistoryID uint, ranges []bronze.GCPComputeSubnetworkSecondaryRange, now time.Time) error {
	tx.Table("bronze_history.gcp_compute_subnetwork_secondary_ranges").
		Where("subnetwork_history_id = ? AND valid_to IS NULL", subnetworkHistoryID).
		Update("valid_to", now)

	for _, sr := range ranges {
		srHist := toSecondaryRangeHistory(&sr, subnetworkHistoryID, now)
		if err := tx.Create(&srHist).Error; err != nil {
			return err
		}
	}
	return nil
}

// Conversion functions: bronze -> bronze_history

func toSubnetworkHistory(subnet *bronze.GCPComputeSubnetwork, now time.Time) bronze_history.GCPComputeSubnetwork {
	return bronze_history.GCPComputeSubnetwork{
		ResourceID:              subnet.ResourceID,
		ValidFrom:               now,
		ValidTo:                 nil,
		Name:                    subnet.Name,
		Description:             subnet.Description,
		SelfLink:                subnet.SelfLink,
		CreationTimestamp:       subnet.CreationTimestamp,
		Network:                 subnet.Network,
		Region:                  subnet.Region,
		IpCidrRange:             subnet.IpCidrRange,
		GatewayAddress:          subnet.GatewayAddress,
		Purpose:                 subnet.Purpose,
		Role:                    subnet.Role,
		PrivateIpGoogleAccess:   subnet.PrivateIpGoogleAccess,
		PrivateIpv6GoogleAccess: subnet.PrivateIpv6GoogleAccess,
		StackType:               subnet.StackType,
		Ipv6AccessType:          subnet.Ipv6AccessType,
		InternalIpv6Prefix:      subnet.InternalIpv6Prefix,
		ExternalIpv6Prefix:      subnet.ExternalIpv6Prefix,
		LogConfigJSON:           subnet.LogConfigJSON,
		Fingerprint:             subnet.Fingerprint,
		ProjectID:               subnet.ProjectID,
		CollectedAt:             subnet.CollectedAt,
	}
}

func toSecondaryRangeHistory(sr *bronze.GCPComputeSubnetworkSecondaryRange, subnetworkHistoryID uint, now time.Time) bronze_history.GCPComputeSubnetworkSecondaryRange {
	return bronze_history.GCPComputeSubnetworkSecondaryRange{
		SubnetworkHistoryID: subnetworkHistoryID,
		ValidFrom:           now,
		ValidTo:             nil,
		RangeName:           sr.RangeName,
		IpCidrRange:         sr.IpCidrRange,
	}
}
