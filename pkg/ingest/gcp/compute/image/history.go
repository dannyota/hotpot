package image

import (
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
	"hotpot/pkg/base/models/bronze_history"
)

// HistoryService handles history tracking for images.
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService creates a new history service.
func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

// CreateHistory creates history records for a new image and all children.
func (h *HistoryService) CreateHistory(tx *gorm.DB, img *bronze.GCPComputeImage, now time.Time) error {
	// Create image history
	imgHist := toImageHistory(img, now)
	if err := tx.Create(&imgHist).Error; err != nil {
		return err
	}

	// Create children history with image_history_id
	return h.createChildrenHistory(tx, imgHist.HistoryID, img, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(tx *gorm.DB, old, new *bronze.GCPComputeImage, diff *ImageDiff, now time.Time) error {
	// Get current image history
	var currentHist bronze_history.GCPComputeImage
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", old.ResourceID).First(&currentHist).Error; err != nil {
		return err
	}

	// If image-level fields changed, close old and create new image history
	if diff.IsChanged {
		// Close old image history
		if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
			return err
		}

		// Create new image history
		imgHist := toImageHistory(new, now)
		if err := tx.Create(&imgHist).Error; err != nil {
			return err
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(tx, currentHist.HistoryID, now); err != nil {
			return err
		}
		return h.createChildrenHistory(tx, imgHist.HistoryID, new, now)
	}

	// Image unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(tx, currentHist.HistoryID, new, diff, now)
}

// CloseHistory closes history records for a deleted image.
func (h *HistoryService) CloseHistory(tx *gorm.DB, resourceID string, now time.Time) error {
	// Get current image history
	var currentHist bronze_history.GCPComputeImage
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", resourceID).First(&currentHist).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // No history to close
		}
		return err
	}

	// Close image history
	if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
		return err
	}

	// Close all children history
	return h.closeChildrenHistory(tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(tx *gorm.DB, imageHistoryID uint, img *bronze.GCPComputeImage, now time.Time) error {
	// Labels
	for _, label := range img.Labels {
		labelHist := toLabelHistory(&label, imageHistoryID, now)
		if err := tx.Create(&labelHist).Error; err != nil {
			return err
		}
	}

	// Licenses
	for _, license := range img.Licenses {
		licHist := toLicenseHistory(&license, imageHistoryID, now)
		if err := tx.Create(&licHist).Error; err != nil {
			return err
		}
	}

	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(tx *gorm.DB, imageHistoryID uint, now time.Time) error {
	// Close labels
	if err := tx.Table("bronze_history.gcp_compute_image_labels").
		Where("image_history_id = ? AND valid_to IS NULL", imageHistoryID).
		Update("valid_to", now).Error; err != nil {
		return err
	}

	// Close licenses
	if err := tx.Table("bronze_history.gcp_compute_image_licenses").
		Where("image_history_id = ? AND valid_to IS NULL", imageHistoryID).
		Update("valid_to", now).Error; err != nil {
		return err
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(tx *gorm.DB, imageHistoryID uint, new *bronze.GCPComputeImage, diff *ImageDiff, now time.Time) error {
	if diff.LabelsDiff.Changed {
		if err := h.updateLabelsHistory(tx, imageHistoryID, new.Labels, now); err != nil {
			return err
		}
	}

	if diff.LicensesDiff.Changed {
		if err := h.updateLicensesHistory(tx, imageHistoryID, new.Licenses, now); err != nil {
			return err
		}
	}

	return nil
}

func (h *HistoryService) updateLabelsHistory(tx *gorm.DB, imageHistoryID uint, labels []bronze.GCPComputeImageLabel, now time.Time) error {
	tx.Table("bronze_history.gcp_compute_image_labels").
		Where("image_history_id = ? AND valid_to IS NULL", imageHistoryID).
		Update("valid_to", now)

	for _, label := range labels {
		labelHist := toLabelHistory(&label, imageHistoryID, now)
		if err := tx.Create(&labelHist).Error; err != nil {
			return err
		}
	}
	return nil
}

func (h *HistoryService) updateLicensesHistory(tx *gorm.DB, imageHistoryID uint, licenses []bronze.GCPComputeImageLicense, now time.Time) error {
	tx.Table("bronze_history.gcp_compute_image_licenses").
		Where("image_history_id = ? AND valid_to IS NULL", imageHistoryID).
		Update("valid_to", now)

	for _, license := range licenses {
		licHist := toLicenseHistory(&license, imageHistoryID, now)
		if err := tx.Create(&licHist).Error; err != nil {
			return err
		}
	}
	return nil
}

// Conversion functions: bronze -> bronze_history

func toImageHistory(img *bronze.GCPComputeImage, now time.Time) bronze_history.GCPComputeImage {
	return bronze_history.GCPComputeImage{
		ResourceID:                       img.ResourceID,
		ValidFrom:                        now,
		ValidTo:                          nil,
		Name:                             img.Name,
		Description:                      img.Description,
		Status:                           img.Status,
		Architecture:                     img.Architecture,
		SelfLink:                         img.SelfLink,
		CreationTimestamp:                img.CreationTimestamp,
		LabelFingerprint:                 img.LabelFingerprint,
		Family:                           img.Family,
		SourceDisk:                       img.SourceDisk,
		SourceDiskId:                     img.SourceDiskId,
		SourceImage:                      img.SourceImage,
		SourceImageId:                    img.SourceImageId,
		SourceSnapshot:                   img.SourceSnapshot,
		SourceSnapshotId:                 img.SourceSnapshotId,
		SourceType:                       img.SourceType,
		DiskSizeGb:                       img.DiskSizeGb,
		ArchiveSizeBytes:                 img.ArchiveSizeBytes,
		SatisfiesPzi:                     img.SatisfiesPzi,
		SatisfiesPzs:                     img.SatisfiesPzs,
		EnableConfidentialCompute:        img.EnableConfidentialCompute,
		ImageEncryptionKeyJSON:           img.ImageEncryptionKeyJSON,
		SourceDiskEncryptionKeyJSON:      img.SourceDiskEncryptionKeyJSON,
		SourceImageEncryptionKeyJSON:     img.SourceImageEncryptionKeyJSON,
		SourceSnapshotEncryptionKeyJSON:  img.SourceSnapshotEncryptionKeyJSON,
		DeprecatedJSON:                   img.DeprecatedJSON,
		GuestOsFeaturesJSON:              img.GuestOsFeaturesJSON,
		ShieldedInstanceInitialStateJSON: img.ShieldedInstanceInitialStateJSON,
		RawDiskJSON:                      img.RawDiskJSON,
		StorageLocationsJSON:             img.StorageLocationsJSON,
		LicenseCodesJSON:                 img.LicenseCodesJSON,
		ProjectID:                        img.ProjectID,
		CollectedAt:                      img.CollectedAt,
	}
}

func toLabelHistory(label *bronze.GCPComputeImageLabel, imageHistoryID uint, now time.Time) bronze_history.GCPComputeImageLabel {
	return bronze_history.GCPComputeImageLabel{
		ImageHistoryID: imageHistoryID,
		ValidFrom:      now,
		ValidTo:        nil,
		Key:            label.Key,
		Value:          label.Value,
	}
}

func toLicenseHistory(license *bronze.GCPComputeImageLicense, imageHistoryID uint, now time.Time) bronze_history.GCPComputeImageLicense {
	return bronze_history.GCPComputeImageLicense{
		ImageHistoryID: imageHistoryID,
		ValidFrom:      now,
		ValidTo:        nil,
		License:        license.License,
	}
}
