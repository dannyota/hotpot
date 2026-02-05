package instance

import (
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
	"hotpot/pkg/base/models/bronze_history"
)

// HistoryService handles history tracking for instances.
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService creates a new history service.
func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

// CreateHistory creates history records for a new instance and all children.
func (h *HistoryService) CreateHistory(tx *gorm.DB, instance *bronze.GCPComputeInstance, now time.Time) error {
	// Create instance history
	instHist := toInstanceHistory(instance, now)
	if err := tx.Create(&instHist).Error; err != nil {
		return err
	}

	// Create children history with instance_history_id
	return h.createChildrenHistory(tx, instHist.HistoryID, instance, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(tx *gorm.DB, old, new *bronze.GCPComputeInstance, diff *InstanceDiff, now time.Time) error {
	// Get current instance history
	var currentHist bronze_history.GCPComputeInstance
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", old.ResourceID).First(&currentHist).Error; err != nil {
		return err
	}

	// If instance-level fields changed, close old and create new instance history
	if diff.IsChanged {
		// Close old instance history
		if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
			return err
		}

		// Create new instance history
		instHist := toInstanceHistory(new, now)
		if err := tx.Create(&instHist).Error; err != nil {
			return err
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(tx, currentHist.HistoryID, now); err != nil {
			return err
		}
		return h.createChildrenHistory(tx, instHist.HistoryID, new, now)
	}

	// Instance unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(tx, currentHist.HistoryID, old, new, diff, now)
}

// CloseHistory closes history records for a deleted instance.
func (h *HistoryService) CloseHistory(tx *gorm.DB, resourceID string, now time.Time) error {
	// Get current instance history
	var currentHist bronze_history.GCPComputeInstance
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", resourceID).First(&currentHist).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // No history to close
		}
		return err
	}

	// Close instance history
	if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
		return err
	}

	// Close all children history
	return h.closeChildrenHistory(tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(tx *gorm.DB, instanceHistoryID uint, instance *bronze.GCPComputeInstance, now time.Time) error {
	// Disks
	for _, disk := range instance.Disks {
		diskHist := toDiskHistory(&disk, instanceHistoryID, now)
		if err := tx.Create(&diskHist).Error; err != nil {
			return err
		}
		// Disk licenses
		for _, lic := range disk.Licenses {
			licHist := toDiskLicenseHistory(&lic, diskHist.HistoryID, now)
			if err := tx.Create(&licHist).Error; err != nil {
				return err
			}
		}
	}

	// NICs
	for _, nic := range instance.NICs {
		nicHist := toNICHistory(&nic, instanceHistoryID, now)
		if err := tx.Create(&nicHist).Error; err != nil {
			return err
		}
		// Access configs
		for _, ac := range nic.AccessConfigs {
			acHist := toAccessConfigHistory(&ac, nicHist.HistoryID, now)
			if err := tx.Create(&acHist).Error; err != nil {
				return err
			}
		}
		// Alias ranges
		for _, ar := range nic.AliasIpRanges {
			arHist := toAliasRangeHistory(&ar, nicHist.HistoryID, now)
			if err := tx.Create(&arHist).Error; err != nil {
				return err
			}
		}
	}

	// Labels
	for _, label := range instance.Labels {
		labelHist := toLabelHistory(&label, instanceHistoryID, now)
		if err := tx.Create(&labelHist).Error; err != nil {
			return err
		}
	}

	// Tags
	for _, tag := range instance.Tags {
		tagHist := toTagHistory(&tag, instanceHistoryID, now)
		if err := tx.Create(&tagHist).Error; err != nil {
			return err
		}
	}

	// Metadata
	for _, meta := range instance.Metadata {
		metaHist := toMetadataHistory(&meta, instanceHistoryID, now)
		if err := tx.Create(&metaHist).Error; err != nil {
			return err
		}
	}

	// Service accounts
	for _, sa := range instance.ServiceAccounts {
		saHist := toServiceAccountHistory(&sa, instanceHistoryID, now)
		if err := tx.Create(&saHist).Error; err != nil {
			return err
		}
	}

	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(tx *gorm.DB, instanceHistoryID uint, now time.Time) error {
	// Close direct children
	tables := []string{
		"bronze_history.gcp_compute_instance_disks",
		"bronze_history.gcp_compute_instance_nics",
		"bronze_history.gcp_compute_instance_labels",
		"bronze_history.gcp_compute_instance_tags",
		"bronze_history.gcp_compute_instance_metadata",
		"bronze_history.gcp_compute_instance_service_accounts",
	}
	for _, table := range tables {
		if err := tx.Table(table).
			Where("instance_history_id = ? AND valid_to IS NULL", instanceHistoryID).
			Update("valid_to", now).Error; err != nil {
			return err
		}
	}

	// Close nested children (disk licenses, nic access configs, alias ranges)
	// Get disk history IDs
	var diskHistIDs []uint
	tx.Table("bronze_history.gcp_compute_instance_disks").
		Where("instance_history_id = ?", instanceHistoryID).
		Pluck("history_id", &diskHistIDs)
	if len(diskHistIDs) > 0 {
		tx.Table("bronze_history.gcp_compute_instance_disk_licenses").
			Where("disk_history_id IN ? AND valid_to IS NULL", diskHistIDs).
			Update("valid_to", now)
	}

	// Get NIC history IDs
	var nicHistIDs []uint
	tx.Table("bronze_history.gcp_compute_instance_nics").
		Where("instance_history_id = ?", instanceHistoryID).
		Pluck("history_id", &nicHistIDs)
	if len(nicHistIDs) > 0 {
		tx.Table("bronze_history.gcp_compute_instance_nic_access_configs").
			Where("nic_history_id IN ? AND valid_to IS NULL", nicHistIDs).
			Update("valid_to", now)
		tx.Table("bronze_history.gcp_compute_instance_nic_alias_ranges").
			Where("nic_history_id IN ? AND valid_to IS NULL", nicHistIDs).
			Update("valid_to", now)
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(tx *gorm.DB, instanceHistoryID uint, old, new *bronze.GCPComputeInstance, diff *InstanceDiff, now time.Time) error {
	// For each child type, if changed: close old + create new
	// If unchanged: no action (still links to same instance_history_id)

	if diff.DisksDiff.Changed {
		if err := h.updateDisksHistory(tx, instanceHistoryID, new.Disks, now); err != nil {
			return err
		}
	}

	if diff.NICsDiff.Changed {
		if err := h.updateNICsHistory(tx, instanceHistoryID, new.NICs, now); err != nil {
			return err
		}
	}

	if diff.LabelsDiff.Changed {
		if err := h.updateLabelsHistory(tx, instanceHistoryID, new.Labels, now); err != nil {
			return err
		}
	}

	if diff.TagsDiff.Changed {
		if err := h.updateTagsHistory(tx, instanceHistoryID, new.Tags, now); err != nil {
			return err
		}
	}

	if diff.MetadataDiff.Changed {
		if err := h.updateMetadataHistory(tx, instanceHistoryID, new.Metadata, now); err != nil {
			return err
		}
	}

	if diff.ServiceAccountsDiff.Changed {
		if err := h.updateServiceAccountsHistory(tx, instanceHistoryID, new.ServiceAccounts, now); err != nil {
			return err
		}
	}

	return nil
}

func (h *HistoryService) updateDisksHistory(tx *gorm.DB, instanceHistoryID uint, disks []bronze.GCPComputeInstanceDisk, now time.Time) error {
	// Close old disk history
	var oldDiskHistIDs []uint
	tx.Table("bronze_history.gcp_compute_instance_disks").
		Where("instance_history_id = ? AND valid_to IS NULL", instanceHistoryID).
		Pluck("history_id", &oldDiskHistIDs)

	if len(oldDiskHistIDs) > 0 {
		tx.Table("bronze_history.gcp_compute_instance_disks").
			Where("history_id IN ?", oldDiskHistIDs).
			Update("valid_to", now)
		tx.Table("bronze_history.gcp_compute_instance_disk_licenses").
			Where("disk_history_id IN ? AND valid_to IS NULL", oldDiskHistIDs).
			Update("valid_to", now)
	}

	// Create new disk history
	for _, disk := range disks {
		diskHist := toDiskHistory(&disk, instanceHistoryID, now)
		if err := tx.Create(&diskHist).Error; err != nil {
			return err
		}
		for _, lic := range disk.Licenses {
			licHist := toDiskLicenseHistory(&lic, diskHist.HistoryID, now)
			if err := tx.Create(&licHist).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *HistoryService) updateNICsHistory(tx *gorm.DB, instanceHistoryID uint, nics []bronze.GCPComputeInstanceNIC, now time.Time) error {
	// Close old NIC history
	var oldNICHistIDs []uint
	tx.Table("bronze_history.gcp_compute_instance_nics").
		Where("instance_history_id = ? AND valid_to IS NULL", instanceHistoryID).
		Pluck("history_id", &oldNICHistIDs)

	if len(oldNICHistIDs) > 0 {
		tx.Table("bronze_history.gcp_compute_instance_nics").
			Where("history_id IN ?", oldNICHistIDs).
			Update("valid_to", now)
		tx.Table("bronze_history.gcp_compute_instance_nic_access_configs").
			Where("nic_history_id IN ? AND valid_to IS NULL", oldNICHistIDs).
			Update("valid_to", now)
		tx.Table("bronze_history.gcp_compute_instance_nic_alias_ranges").
			Where("nic_history_id IN ? AND valid_to IS NULL", oldNICHistIDs).
			Update("valid_to", now)
	}

	// Create new NIC history
	for _, nic := range nics {
		nicHist := toNICHistory(&nic, instanceHistoryID, now)
		if err := tx.Create(&nicHist).Error; err != nil {
			return err
		}
		for _, ac := range nic.AccessConfigs {
			acHist := toAccessConfigHistory(&ac, nicHist.HistoryID, now)
			if err := tx.Create(&acHist).Error; err != nil {
				return err
			}
		}
		for _, ar := range nic.AliasIpRanges {
			arHist := toAliasRangeHistory(&ar, nicHist.HistoryID, now)
			if err := tx.Create(&arHist).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *HistoryService) updateLabelsHistory(tx *gorm.DB, instanceHistoryID uint, labels []bronze.GCPComputeInstanceLabel, now time.Time) error {
	tx.Table("bronze_history.gcp_compute_instance_labels").
		Where("instance_history_id = ? AND valid_to IS NULL", instanceHistoryID).
		Update("valid_to", now)

	for _, label := range labels {
		labelHist := toLabelHistory(&label, instanceHistoryID, now)
		if err := tx.Create(&labelHist).Error; err != nil {
			return err
		}
	}
	return nil
}

func (h *HistoryService) updateTagsHistory(tx *gorm.DB, instanceHistoryID uint, tags []bronze.GCPComputeInstanceTag, now time.Time) error {
	tx.Table("bronze_history.gcp_compute_instance_tags").
		Where("instance_history_id = ? AND valid_to IS NULL", instanceHistoryID).
		Update("valid_to", now)

	for _, tag := range tags {
		tagHist := toTagHistory(&tag, instanceHistoryID, now)
		if err := tx.Create(&tagHist).Error; err != nil {
			return err
		}
	}
	return nil
}

func (h *HistoryService) updateMetadataHistory(tx *gorm.DB, instanceHistoryID uint, metadata []bronze.GCPComputeInstanceMetadata, now time.Time) error {
	tx.Table("bronze_history.gcp_compute_instance_metadata").
		Where("instance_history_id = ? AND valid_to IS NULL", instanceHistoryID).
		Update("valid_to", now)

	for _, meta := range metadata {
		metaHist := toMetadataHistory(&meta, instanceHistoryID, now)
		if err := tx.Create(&metaHist).Error; err != nil {
			return err
		}
	}
	return nil
}

func (h *HistoryService) updateServiceAccountsHistory(tx *gorm.DB, instanceHistoryID uint, sas []bronze.GCPComputeInstanceServiceAccount, now time.Time) error {
	tx.Table("bronze_history.gcp_compute_instance_service_accounts").
		Where("instance_history_id = ? AND valid_to IS NULL", instanceHistoryID).
		Update("valid_to", now)

	for _, sa := range sas {
		saHist := toServiceAccountHistory(&sa, instanceHistoryID, now)
		if err := tx.Create(&saHist).Error; err != nil {
			return err
		}
	}
	return nil
}

// Conversion functions: bronze -> bronze_history

func toInstanceHistory(inst *bronze.GCPComputeInstance, now time.Time) bronze_history.GCPComputeInstance {
	return bronze_history.GCPComputeInstance{
		ResourceID:             inst.ResourceID,
		ValidFrom:              now,
		ValidTo:                nil,
		Name:                   inst.Name,
		Zone:                   inst.Zone,
		MachineType:            inst.MachineType,
		Status:                 inst.Status,
		StatusMessage:          inst.StatusMessage,
		CpuPlatform:            inst.CpuPlatform,
		Hostname:               inst.Hostname,
		Description:            inst.Description,
		CreationTimestamp:      inst.CreationTimestamp,
		LastStartTimestamp:     inst.LastStartTimestamp,
		LastStopTimestamp:      inst.LastStopTimestamp,
		LastSuspendedTimestamp: inst.LastSuspendedTimestamp,
		DeletionProtection:     inst.DeletionProtection,
		CanIpForward:           inst.CanIpForward,
		SelfLink:               inst.SelfLink,
		SchedulingJSON:         inst.SchedulingJSON,
		ProjectID:              inst.ProjectID,
		CollectedAt:            inst.CollectedAt,
	}
}

func toDiskHistory(disk *bronze.GCPComputeInstanceDisk, instanceHistoryID uint, now time.Time) bronze_history.GCPComputeInstanceDisk {
	return bronze_history.GCPComputeInstanceDisk{
		InstanceHistoryID:     instanceHistoryID,
		ValidFrom:             now,
		ValidTo:               nil,
		Source:                disk.Source,
		DeviceName:            disk.DeviceName,
		Index:                 disk.Index,
		Boot:                  disk.Boot,
		AutoDelete:            disk.AutoDelete,
		Mode:                  disk.Mode,
		Interface:             disk.Interface,
		Type:                  disk.Type,
		DiskSizeGb:            disk.DiskSizeGb,
		DiskEncryptionKeyJSON: disk.DiskEncryptionKeyJSON,
		InitializeParamsJSON:  disk.InitializeParamsJSON,
	}
}

func toDiskLicenseHistory(lic *bronze.GCPComputeInstanceDiskLicense, diskHistoryID uint, now time.Time) bronze_history.GCPComputeInstanceDiskLicense {
	return bronze_history.GCPComputeInstanceDiskLicense{
		DiskHistoryID: diskHistoryID,
		ValidFrom:     now,
		ValidTo:       nil,
		License:       lic.License,
	}
}

func toNICHistory(nic *bronze.GCPComputeInstanceNIC, instanceHistoryID uint, now time.Time) bronze_history.GCPComputeInstanceNIC {
	return bronze_history.GCPComputeInstanceNIC{
		InstanceHistoryID: instanceHistoryID,
		ValidFrom:         now,
		ValidTo:           nil,
		Name:              nic.Name,
		Network:           nic.Network,
		Subnetwork:        nic.Subnetwork,
		NetworkIP:         nic.NetworkIP,
		StackType:         nic.StackType,
		NicType:           nic.NicType,
	}
}

func toAccessConfigHistory(ac *bronze.GCPComputeInstanceNICAccessConfig, nicHistoryID uint, now time.Time) bronze_history.GCPComputeInstanceNICAccessConfig {
	return bronze_history.GCPComputeInstanceNICAccessConfig{
		NICHistoryID: nicHistoryID,
		ValidFrom:    now,
		ValidTo:      nil,
		Type:         ac.Type,
		Name:         ac.Name,
		NatIP:        ac.NatIP,
		NetworkTier:  ac.NetworkTier,
	}
}

func toAliasRangeHistory(ar *bronze.GCPComputeInstanceNICAliasRange, nicHistoryID uint, now time.Time) bronze_history.GCPComputeInstanceNICAliasRange {
	return bronze_history.GCPComputeInstanceNICAliasRange{
		NICHistoryID:        nicHistoryID,
		ValidFrom:           now,
		ValidTo:             nil,
		IpCidrRange:         ar.IpCidrRange,
		SubnetworkRangeName: ar.SubnetworkRangeName,
	}
}

func toLabelHistory(label *bronze.GCPComputeInstanceLabel, instanceHistoryID uint, now time.Time) bronze_history.GCPComputeInstanceLabel {
	return bronze_history.GCPComputeInstanceLabel{
		InstanceHistoryID: instanceHistoryID,
		ValidFrom:         now,
		ValidTo:           nil,
		Key:               label.Key,
		Value:             label.Value,
	}
}

func toTagHistory(tag *bronze.GCPComputeInstanceTag, instanceHistoryID uint, now time.Time) bronze_history.GCPComputeInstanceTag {
	return bronze_history.GCPComputeInstanceTag{
		InstanceHistoryID: instanceHistoryID,
		ValidFrom:         now,
		ValidTo:           nil,
		Tag:               tag.Tag,
	}
}

func toMetadataHistory(meta *bronze.GCPComputeInstanceMetadata, instanceHistoryID uint, now time.Time) bronze_history.GCPComputeInstanceMetadata {
	return bronze_history.GCPComputeInstanceMetadata{
		InstanceHistoryID: instanceHistoryID,
		ValidFrom:         now,
		ValidTo:           nil,
		Key:               meta.Key,
		Value:             meta.Value,
	}
}

func toServiceAccountHistory(sa *bronze.GCPComputeInstanceServiceAccount, instanceHistoryID uint, now time.Time) bronze_history.GCPComputeInstanceServiceAccount {
	return bronze_history.GCPComputeInstanceServiceAccount{
		InstanceHistoryID: instanceHistoryID,
		ValidFrom:         now,
		ValidTo:           nil,
		Email:             sa.Email,
		ScopesJSON:        sa.ScopesJSON,
	}
}
