package disk

import (
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
	"hotpot/pkg/base/models/bronze_history"
)

// HistoryService handles history tracking for disks.
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService creates a new history service.
func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

// CreateHistory creates history records for a new disk and all children.
func (h *HistoryService) CreateHistory(tx *gorm.DB, disk *bronze.GCPComputeDisk, now time.Time) error {
	// Create disk history
	diskHist := toDiskHistory(disk, now)
	if err := tx.Create(&diskHist).Error; err != nil {
		return err
	}

	// Create children history with disk_history_id
	return h.createChildrenHistory(tx, diskHist.HistoryID, disk, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(tx *gorm.DB, old, new *bronze.GCPComputeDisk, diff *DiskDiff, now time.Time) error {
	// Get current disk history
	var currentHist bronze_history.GCPComputeDisk
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", old.ResourceID).First(&currentHist).Error; err != nil {
		return err
	}

	// If disk-level fields changed, close old and create new disk history
	if diff.IsChanged {
		// Close old disk history
		if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
			return err
		}

		// Create new disk history
		diskHist := toDiskHistory(new, now)
		if err := tx.Create(&diskHist).Error; err != nil {
			return err
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(tx, currentHist.HistoryID, now); err != nil {
			return err
		}
		return h.createChildrenHistory(tx, diskHist.HistoryID, new, now)
	}

	// Disk unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(tx, currentHist.HistoryID, new, diff, now)
}

// CloseHistory closes history records for a deleted disk.
func (h *HistoryService) CloseHistory(tx *gorm.DB, resourceID string, now time.Time) error {
	// Get current disk history
	var currentHist bronze_history.GCPComputeDisk
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", resourceID).First(&currentHist).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // No history to close
		}
		return err
	}

	// Close disk history
	if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
		return err
	}

	// Close all children history
	return h.closeChildrenHistory(tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(tx *gorm.DB, diskHistoryID uint, disk *bronze.GCPComputeDisk, now time.Time) error {
	// Labels
	for _, label := range disk.Labels {
		labelHist := toLabelHistory(&label, diskHistoryID, now)
		if err := tx.Create(&labelHist).Error; err != nil {
			return err
		}
	}

	// Licenses
	for _, license := range disk.Licenses {
		licHist := toLicenseHistory(&license, diskHistoryID, now)
		if err := tx.Create(&licHist).Error; err != nil {
			return err
		}
	}

	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(tx *gorm.DB, diskHistoryID uint, now time.Time) error {
	// Close labels
	if err := tx.Table("bronze_history.gcp_compute_disk_labels").
		Where("disk_history_id = ? AND valid_to IS NULL", diskHistoryID).
		Update("valid_to", now).Error; err != nil {
		return err
	}

	// Close licenses
	if err := tx.Table("bronze_history.gcp_compute_disk_licenses").
		Where("disk_history_id = ? AND valid_to IS NULL", diskHistoryID).
		Update("valid_to", now).Error; err != nil {
		return err
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(tx *gorm.DB, diskHistoryID uint, new *bronze.GCPComputeDisk, diff *DiskDiff, now time.Time) error {
	if diff.LabelsDiff.Changed {
		if err := h.updateLabelsHistory(tx, diskHistoryID, new.Labels, now); err != nil {
			return err
		}
	}

	if diff.LicensesDiff.Changed {
		if err := h.updateLicensesHistory(tx, diskHistoryID, new.Licenses, now); err != nil {
			return err
		}
	}

	return nil
}

func (h *HistoryService) updateLabelsHistory(tx *gorm.DB, diskHistoryID uint, labels []bronze.GCPComputeDiskLabel, now time.Time) error {
	tx.Table("bronze_history.gcp_compute_disk_labels").
		Where("disk_history_id = ? AND valid_to IS NULL", diskHistoryID).
		Update("valid_to", now)

	for _, label := range labels {
		labelHist := toLabelHistory(&label, diskHistoryID, now)
		if err := tx.Create(&labelHist).Error; err != nil {
			return err
		}
	}
	return nil
}

func (h *HistoryService) updateLicensesHistory(tx *gorm.DB, diskHistoryID uint, licenses []bronze.GCPComputeDiskLicense, now time.Time) error {
	tx.Table("bronze_history.gcp_compute_disk_licenses").
		Where("disk_history_id = ? AND valid_to IS NULL", diskHistoryID).
		Update("valid_to", now)

	for _, license := range licenses {
		licHist := toLicenseHistory(&license, diskHistoryID, now)
		if err := tx.Create(&licHist).Error; err != nil {
			return err
		}
	}
	return nil
}

// Conversion functions: bronze -> bronze_history

func toDiskHistory(disk *bronze.GCPComputeDisk, now time.Time) bronze_history.GCPComputeDisk {
	return bronze_history.GCPComputeDisk{
		ResourceID:                disk.ResourceID,
		ValidFrom:                 now,
		ValidTo:                   nil,
		Name:                      disk.Name,
		Description:               disk.Description,
		Zone:                      disk.Zone,
		Region:                    disk.Region,
		Type:                      disk.Type,
		Status:                    disk.Status,
		SizeGb:                    disk.SizeGb,
		Architecture:              disk.Architecture,
		SelfLink:                  disk.SelfLink,
		CreationTimestamp:         disk.CreationTimestamp,
		LastAttachTimestamp:       disk.LastAttachTimestamp,
		LastDetachTimestamp:       disk.LastDetachTimestamp,
		SourceImage:               disk.SourceImage,
		SourceImageId:             disk.SourceImageId,
		SourceSnapshot:            disk.SourceSnapshot,
		SourceSnapshotId:          disk.SourceSnapshotId,
		SourceDisk:                disk.SourceDisk,
		SourceDiskId:              disk.SourceDiskId,
		ProvisionedIops:           disk.ProvisionedIops,
		ProvisionedThroughput:     disk.ProvisionedThroughput,
		PhysicalBlockSizeBytes:    disk.PhysicalBlockSizeBytes,
		EnableConfidentialCompute: disk.EnableConfidentialCompute,
		DiskEncryptionKeyJSON:     disk.DiskEncryptionKeyJSON,
		UsersJSON:                 disk.UsersJSON,
		ReplicaZonesJSON:          disk.ReplicaZonesJSON,
		ResourcePoliciesJSON:      disk.ResourcePoliciesJSON,
		GuestOsFeaturesJSON:       disk.GuestOsFeaturesJSON,
		ProjectID:                 disk.ProjectID,
		CollectedAt:               disk.CollectedAt,
	}
}

func toLabelHistory(label *bronze.GCPComputeDiskLabel, diskHistoryID uint, now time.Time) bronze_history.GCPComputeDiskLabel {
	return bronze_history.GCPComputeDiskLabel{
		DiskHistoryID: diskHistoryID,
		ValidFrom:     now,
		ValidTo:       nil,
		Key:           label.Key,
		Value:         label.Value,
	}
}

func toLicenseHistory(license *bronze.GCPComputeDiskLicense, diskHistoryID uint, now time.Time) bronze_history.GCPComputeDiskLicense {
	return bronze_history.GCPComputeDiskLicense{
		DiskHistoryID: diskHistoryID,
		ValidFrom:     now,
		ValidTo:       nil,
		License:       license.License,
	}
}
