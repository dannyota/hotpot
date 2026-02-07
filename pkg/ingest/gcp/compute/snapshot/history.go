package snapshot

import (
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
	"hotpot/pkg/base/models/bronze_history"
)

// HistoryService handles history tracking for snapshots.
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService creates a new history service.
func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

// CreateHistory creates history records for a new snapshot and all children.
func (h *HistoryService) CreateHistory(tx *gorm.DB, snap *bronze.GCPComputeSnapshot, now time.Time) error {
	// Create snapshot history
	snapHist := toSnapshotHistory(snap, now)
	if err := tx.Create(&snapHist).Error; err != nil {
		return err
	}

	// Create children history with snapshot_history_id
	return h.createChildrenHistory(tx, snapHist.HistoryID, snap, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(tx *gorm.DB, old, new *bronze.GCPComputeSnapshot, diff *SnapshotDiff, now time.Time) error {
	// Get current snapshot history
	var currentHist bronze_history.GCPComputeSnapshot
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", old.ResourceID).First(&currentHist).Error; err != nil {
		return err
	}

	// If snapshot-level fields changed, close old and create new snapshot history
	if diff.IsChanged {
		// Close old snapshot history
		if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
			return err
		}

		// Create new snapshot history
		snapHist := toSnapshotHistory(new, now)
		if err := tx.Create(&snapHist).Error; err != nil {
			return err
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(tx, currentHist.HistoryID, now); err != nil {
			return err
		}
		return h.createChildrenHistory(tx, snapHist.HistoryID, new, now)
	}

	// Snapshot unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(tx, currentHist.HistoryID, new, diff, now)
}

// CloseHistory closes history records for a deleted snapshot.
func (h *HistoryService) CloseHistory(tx *gorm.DB, resourceID string, now time.Time) error {
	// Get current snapshot history
	var currentHist bronze_history.GCPComputeSnapshot
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", resourceID).First(&currentHist).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // No history to close
		}
		return err
	}

	// Close snapshot history
	if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
		return err
	}

	// Close all children history
	return h.closeChildrenHistory(tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(tx *gorm.DB, snapshotHistoryID uint, snap *bronze.GCPComputeSnapshot, now time.Time) error {
	// Labels
	for _, label := range snap.Labels {
		labelHist := toLabelHistory(&label, snapshotHistoryID, now)
		if err := tx.Create(&labelHist).Error; err != nil {
			return err
		}
	}

	// Licenses
	for _, license := range snap.Licenses {
		licHist := toLicenseHistory(&license, snapshotHistoryID, now)
		if err := tx.Create(&licHist).Error; err != nil {
			return err
		}
	}

	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(tx *gorm.DB, snapshotHistoryID uint, now time.Time) error {
	// Close labels
	if err := tx.Table("bronze_history.gcp_compute_snapshot_labels").
		Where("snapshot_history_id = ? AND valid_to IS NULL", snapshotHistoryID).
		Update("valid_to", now).Error; err != nil {
		return err
	}

	// Close licenses
	if err := tx.Table("bronze_history.gcp_compute_snapshot_licenses").
		Where("snapshot_history_id = ? AND valid_to IS NULL", snapshotHistoryID).
		Update("valid_to", now).Error; err != nil {
		return err
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(tx *gorm.DB, snapshotHistoryID uint, new *bronze.GCPComputeSnapshot, diff *SnapshotDiff, now time.Time) error {
	if diff.LabelsDiff.Changed {
		if err := h.updateLabelsHistory(tx, snapshotHistoryID, new.Labels, now); err != nil {
			return err
		}
	}

	if diff.LicensesDiff.Changed {
		if err := h.updateLicensesHistory(tx, snapshotHistoryID, new.Licenses, now); err != nil {
			return err
		}
	}

	return nil
}

func (h *HistoryService) updateLabelsHistory(tx *gorm.DB, snapshotHistoryID uint, labels []bronze.GCPComputeSnapshotLabel, now time.Time) error {
	tx.Table("bronze_history.gcp_compute_snapshot_labels").
		Where("snapshot_history_id = ? AND valid_to IS NULL", snapshotHistoryID).
		Update("valid_to", now)

	for _, label := range labels {
		labelHist := toLabelHistory(&label, snapshotHistoryID, now)
		if err := tx.Create(&labelHist).Error; err != nil {
			return err
		}
	}
	return nil
}

func (h *HistoryService) updateLicensesHistory(tx *gorm.DB, snapshotHistoryID uint, licenses []bronze.GCPComputeSnapshotLicense, now time.Time) error {
	tx.Table("bronze_history.gcp_compute_snapshot_licenses").
		Where("snapshot_history_id = ? AND valid_to IS NULL", snapshotHistoryID).
		Update("valid_to", now)

	for _, license := range licenses {
		licHist := toLicenseHistory(&license, snapshotHistoryID, now)
		if err := tx.Create(&licHist).Error; err != nil {
			return err
		}
	}
	return nil
}

// Conversion functions: bronze -> bronze_history

func toSnapshotHistory(snap *bronze.GCPComputeSnapshot, now time.Time) bronze_history.GCPComputeSnapshot {
	return bronze_history.GCPComputeSnapshot{
		ResourceID:                      snap.ResourceID,
		ValidFrom:                       now,
		ValidTo:                         nil,
		Name:                            snap.Name,
		Description:                     snap.Description,
		Status:                          snap.Status,
		DiskSizeGb:                      snap.DiskSizeGb,
		StorageBytes:                    snap.StorageBytes,
		StorageBytesStatus:              snap.StorageBytesStatus,
		DownloadBytes:                   snap.DownloadBytes,
		SnapshotType:                    snap.SnapshotType,
		Architecture:                    snap.Architecture,
		SelfLink:                        snap.SelfLink,
		CreationTimestamp:               snap.CreationTimestamp,
		LabelFingerprint:                snap.LabelFingerprint,
		SourceDisk:                      snap.SourceDisk,
		SourceDiskId:                    snap.SourceDiskId,
		SourceDiskForRecoveryCheckpoint: snap.SourceDiskForRecoveryCheckpoint,
		AutoCreated:                     snap.AutoCreated,
		SatisfiesPzi:                    snap.SatisfiesPzi,
		SatisfiesPzs:                    snap.SatisfiesPzs,
		EnableConfidentialCompute:       snap.EnableConfidentialCompute,
		SnapshotEncryptionKeyJSON:       snap.SnapshotEncryptionKeyJSON,
		SourceDiskEncryptionKeyJSON:     snap.SourceDiskEncryptionKeyJSON,
		GuestOsFeaturesJSON:             snap.GuestOsFeaturesJSON,
		StorageLocationsJSON:            snap.StorageLocationsJSON,
		ProjectID:                       snap.ProjectID,
		CollectedAt:                     snap.CollectedAt,
	}
}

func toLabelHistory(label *bronze.GCPComputeSnapshotLabel, snapshotHistoryID uint, now time.Time) bronze_history.GCPComputeSnapshotLabel {
	return bronze_history.GCPComputeSnapshotLabel{
		SnapshotHistoryID: snapshotHistoryID,
		ValidFrom:         now,
		ValidTo:           nil,
		Key:               label.Key,
		Value:             label.Value,
	}
}

func toLicenseHistory(license *bronze.GCPComputeSnapshotLicense, snapshotHistoryID uint, now time.Time) bronze_history.GCPComputeSnapshotLicense {
	return bronze_history.GCPComputeSnapshotLicense{
		SnapshotHistoryID: snapshotHistoryID,
		ValidFrom:         now,
		ValidTo:           nil,
		License:           license.License,
	}
}
