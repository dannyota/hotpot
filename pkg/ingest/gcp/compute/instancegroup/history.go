package instancegroup

import (
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
	"hotpot/pkg/base/models/bronze_history"
)

// HistoryService handles history tracking for instance groups.
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService creates a new history service.
func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

// CreateHistory creates history records for a new instance group and all children.
func (h *HistoryService) CreateHistory(tx *gorm.DB, group *bronze.GCPComputeInstanceGroup, now time.Time) error {
	// Create group history
	groupHist := toGroupHistory(group, now)
	if err := tx.Create(&groupHist).Error; err != nil {
		return err
	}

	// Create children history with group_history_id
	return h.createChildrenHistory(tx, groupHist.HistoryID, group, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(tx *gorm.DB, old, new *bronze.GCPComputeInstanceGroup, diff *InstanceGroupDiff, now time.Time) error {
	// Get current group history
	var currentHist bronze_history.GCPComputeInstanceGroup
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", old.ResourceID).First(&currentHist).Error; err != nil {
		return err
	}

	// If group-level fields changed, close old and create new group history
	if diff.IsChanged {
		// Close old group history
		if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
			return err
		}

		// Create new group history
		groupHist := toGroupHistory(new, now)
		if err := tx.Create(&groupHist).Error; err != nil {
			return err
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(tx, currentHist.HistoryID, now); err != nil {
			return err
		}
		return h.createChildrenHistory(tx, groupHist.HistoryID, new, now)
	}

	// Group unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(tx, currentHist.HistoryID, new, diff, now)
}

// CloseHistory closes history records for a deleted instance group.
func (h *HistoryService) CloseHistory(tx *gorm.DB, resourceID string, now time.Time) error {
	// Get current group history
	var currentHist bronze_history.GCPComputeInstanceGroup
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", resourceID).First(&currentHist).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // No history to close
		}
		return err
	}

	// Close group history
	if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
		return err
	}

	// Close all children history
	return h.closeChildrenHistory(tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(tx *gorm.DB, groupHistoryID uint, group *bronze.GCPComputeInstanceGroup, now time.Time) error {
	// Named ports
	for _, port := range group.NamedPorts {
		portHist := toNamedPortHistory(&port, groupHistoryID, now)
		if err := tx.Create(&portHist).Error; err != nil {
			return err
		}
	}

	// Members
	for _, member := range group.Members {
		memberHist := toMemberHistory(&member, groupHistoryID, now)
		if err := tx.Create(&memberHist).Error; err != nil {
			return err
		}
	}

	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(tx *gorm.DB, groupHistoryID uint, now time.Time) error {
	// Close named ports
	if err := tx.Table("bronze_history.gcp_compute_instance_group_named_ports").
		Where("group_history_id = ? AND valid_to IS NULL", groupHistoryID).
		Update("valid_to", now).Error; err != nil {
		return err
	}

	// Close members
	if err := tx.Table("bronze_history.gcp_compute_instance_group_members").
		Where("group_history_id = ? AND valid_to IS NULL", groupHistoryID).
		Update("valid_to", now).Error; err != nil {
		return err
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(tx *gorm.DB, groupHistoryID uint, new *bronze.GCPComputeInstanceGroup, diff *InstanceGroupDiff, now time.Time) error {
	if diff.NamedPortsDiff.Changed {
		if err := h.updateNamedPortsHistory(tx, groupHistoryID, new.NamedPorts, now); err != nil {
			return err
		}
	}

	if diff.MembersDiff.Changed {
		if err := h.updateMembersHistory(tx, groupHistoryID, new.Members, now); err != nil {
			return err
		}
	}

	return nil
}

func (h *HistoryService) updateNamedPortsHistory(tx *gorm.DB, groupHistoryID uint, ports []bronze.GCPComputeInstanceGroupNamedPort, now time.Time) error {
	if err := tx.Table("bronze_history.gcp_compute_instance_group_named_ports").
		Where("group_history_id = ? AND valid_to IS NULL", groupHistoryID).
		Update("valid_to", now).Error; err != nil {
		return err
	}

	for _, port := range ports {
		portHist := toNamedPortHistory(&port, groupHistoryID, now)
		if err := tx.Create(&portHist).Error; err != nil {
			return err
		}
	}
	return nil
}

func (h *HistoryService) updateMembersHistory(tx *gorm.DB, groupHistoryID uint, members []bronze.GCPComputeInstanceGroupMember, now time.Time) error {
	if err := tx.Table("bronze_history.gcp_compute_instance_group_members").
		Where("group_history_id = ? AND valid_to IS NULL", groupHistoryID).
		Update("valid_to", now).Error; err != nil {
		return err
	}

	for _, member := range members {
		memberHist := toMemberHistory(&member, groupHistoryID, now)
		if err := tx.Create(&memberHist).Error; err != nil {
			return err
		}
	}
	return nil
}

// Conversion functions: bronze -> bronze_history

func toGroupHistory(group *bronze.GCPComputeInstanceGroup, now time.Time) bronze_history.GCPComputeInstanceGroup {
	return bronze_history.GCPComputeInstanceGroup{
		ResourceID:        group.ResourceID,
		ValidFrom:         now,
		ValidTo:           nil,
		Name:              group.Name,
		Description:       group.Description,
		Zone:              group.Zone,
		Network:           group.Network,
		Subnetwork:        group.Subnetwork,
		Size:              group.Size,
		SelfLink:          group.SelfLink,
		CreationTimestamp: group.CreationTimestamp,
		Fingerprint:       group.Fingerprint,
		ProjectID:         group.ProjectID,
		CollectedAt:       group.CollectedAt,
	}
}

func toNamedPortHistory(port *bronze.GCPComputeInstanceGroupNamedPort, groupHistoryID uint, now time.Time) bronze_history.GCPComputeInstanceGroupNamedPort {
	return bronze_history.GCPComputeInstanceGroupNamedPort{
		GroupHistoryID: groupHistoryID,
		ValidFrom:      now,
		ValidTo:        nil,
		Name:           port.Name,
		Port:           port.Port,
	}
}

func toMemberHistory(member *bronze.GCPComputeInstanceGroupMember, groupHistoryID uint, now time.Time) bronze_history.GCPComputeInstanceGroupMember {
	return bronze_history.GCPComputeInstanceGroupMember{
		GroupHistoryID: groupHistoryID,
		ValidFrom:      now,
		ValidTo:        nil,
		InstanceURL:    member.InstanceURL,
		InstanceName:   member.InstanceName,
		Status:         member.Status,
	}
}
