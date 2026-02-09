package image

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcomputeimage"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcomputeimagelabel"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcomputeimagelicense"
)

// HistoryService handles history tracking for images.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new image and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, imageData *ImageData, now time.Time) error {
	// Create image history
	imgHistCreate := tx.BronzeHistoryGCPComputeImage.Create().
		SetResourceID(imageData.ID).
		SetValidFrom(now).
		SetCollectedAt(imageData.CollectedAt).
		SetFirstCollectedAt(imageData.CollectedAt).
		SetName(imageData.Name).
		SetDescription(imageData.Description).
		SetStatus(imageData.Status).
		SetArchitecture(imageData.Architecture).
		SetSelfLink(imageData.SelfLink).
		SetCreationTimestamp(imageData.CreationTimestamp).
		SetLabelFingerprint(imageData.LabelFingerprint).
		SetFamily(imageData.Family).
		SetSourceDisk(imageData.SourceDisk).
		SetSourceDiskID(imageData.SourceDiskId).
		SetSourceImage(imageData.SourceImage).
		SetSourceImageID(imageData.SourceImageId).
		SetSourceSnapshot(imageData.SourceSnapshot).
		SetSourceSnapshotID(imageData.SourceSnapshotId).
		SetSourceType(imageData.SourceType).
		SetDiskSizeGB(imageData.DiskSizeGb).
		SetArchiveSizeBytes(imageData.ArchiveSizeBytes).
		SetSatisfiesPzi(imageData.SatisfiesPzi).
		SetSatisfiesPzs(imageData.SatisfiesPzs).
		SetEnableConfidentialCompute(imageData.EnableConfidentialCompute).
		SetProjectID(imageData.ProjectID)

	if imageData.ImageEncryptionKeyJSON != nil {
		imgHistCreate.SetImageEncryptionKeyJSON(imageData.ImageEncryptionKeyJSON)
	}
	if imageData.SourceDiskEncryptionKeyJSON != nil {
		imgHistCreate.SetSourceDiskEncryptionKeyJSON(imageData.SourceDiskEncryptionKeyJSON)
	}
	if imageData.SourceImageEncryptionKeyJSON != nil {
		imgHistCreate.SetSourceImageEncryptionKeyJSON(imageData.SourceImageEncryptionKeyJSON)
	}
	if imageData.SourceSnapshotEncryptionKeyJSON != nil {
		imgHistCreate.SetSourceSnapshotEncryptionKeyJSON(imageData.SourceSnapshotEncryptionKeyJSON)
	}
	if imageData.DeprecatedJSON != nil {
		imgHistCreate.SetDeprecatedJSON(imageData.DeprecatedJSON)
	}
	if imageData.GuestOsFeaturesJSON != nil {
		imgHistCreate.SetGuestOsFeaturesJSON(imageData.GuestOsFeaturesJSON)
	}
	if imageData.ShieldedInstanceInitialStateJSON != nil {
		imgHistCreate.SetShieldedInstanceInitialStateJSON(imageData.ShieldedInstanceInitialStateJSON)
	}
	if imageData.RawDiskJSON != nil {
		imgHistCreate.SetRawDiskJSON(imageData.RawDiskJSON)
	}
	if imageData.StorageLocationsJSON != nil {
		imgHistCreate.SetStorageLocationsJSON(imageData.StorageLocationsJSON)
	}
	if imageData.LicenseCodesJSON != nil {
		imgHistCreate.SetLicenseCodesJSON(imageData.LicenseCodesJSON)
	}

	imgHist, err := imgHistCreate.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create image history: %w", err)
	}

	// Create children history with image_history_id
	return h.createChildrenHistory(ctx, tx, imgHist.HistoryID, imageData, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeImage, new *ImageData, diff *ImageDiff, now time.Time) error {
	// Get current image history
	currentHist, err := tx.BronzeHistoryGCPComputeImage.Query().
		Where(
			bronzehistorygcpcomputeimage.ResourceID(old.ID),
			bronzehistorygcpcomputeimage.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current image history: %w", err)
	}

	// If image-level fields changed, close old and create new image history
	if diff.IsChanged {
		// Close old image history
		if err := tx.BronzeHistoryGCPComputeImage.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to close image history: %w", err)
		}

		// Create new image history
		imgHistCreate := tx.BronzeHistoryGCPComputeImage.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetDescription(new.Description).
			SetStatus(new.Status).
			SetArchitecture(new.Architecture).
			SetSelfLink(new.SelfLink).
			SetCreationTimestamp(new.CreationTimestamp).
			SetLabelFingerprint(new.LabelFingerprint).
			SetFamily(new.Family).
			SetSourceDisk(new.SourceDisk).
			SetSourceDiskID(new.SourceDiskId).
			SetSourceImage(new.SourceImage).
			SetSourceImageID(new.SourceImageId).
			SetSourceSnapshot(new.SourceSnapshot).
			SetSourceSnapshotID(new.SourceSnapshotId).
			SetSourceType(new.SourceType).
			SetDiskSizeGB(new.DiskSizeGb).
			SetArchiveSizeBytes(new.ArchiveSizeBytes).
			SetSatisfiesPzi(new.SatisfiesPzi).
			SetSatisfiesPzs(new.SatisfiesPzs).
			SetEnableConfidentialCompute(new.EnableConfidentialCompute).
			SetProjectID(new.ProjectID)

		if new.ImageEncryptionKeyJSON != nil {
			imgHistCreate.SetImageEncryptionKeyJSON(new.ImageEncryptionKeyJSON)
		}
		if new.SourceDiskEncryptionKeyJSON != nil {
			imgHistCreate.SetSourceDiskEncryptionKeyJSON(new.SourceDiskEncryptionKeyJSON)
		}
		if new.SourceImageEncryptionKeyJSON != nil {
			imgHistCreate.SetSourceImageEncryptionKeyJSON(new.SourceImageEncryptionKeyJSON)
		}
		if new.SourceSnapshotEncryptionKeyJSON != nil {
			imgHistCreate.SetSourceSnapshotEncryptionKeyJSON(new.SourceSnapshotEncryptionKeyJSON)
		}
		if new.DeprecatedJSON != nil {
			imgHistCreate.SetDeprecatedJSON(new.DeprecatedJSON)
		}
		if new.GuestOsFeaturesJSON != nil {
			imgHistCreate.SetGuestOsFeaturesJSON(new.GuestOsFeaturesJSON)
		}
		if new.ShieldedInstanceInitialStateJSON != nil {
			imgHistCreate.SetShieldedInstanceInitialStateJSON(new.ShieldedInstanceInitialStateJSON)
		}
		if new.RawDiskJSON != nil {
			imgHistCreate.SetRawDiskJSON(new.RawDiskJSON)
		}
		if new.StorageLocationsJSON != nil {
			imgHistCreate.SetStorageLocationsJSON(new.StorageLocationsJSON)
		}
		if new.LicenseCodesJSON != nil {
			imgHistCreate.SetLicenseCodesJSON(new.LicenseCodesJSON)
		}

		imgHist, err := imgHistCreate.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create image history: %w", err)
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(ctx, tx, currentHist.HistoryID, now); err != nil {
			return fmt.Errorf("failed to close children history: %w", err)
		}
		return h.createChildrenHistory(ctx, tx, imgHist.HistoryID, new, now)
	}

	// Image unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(ctx, tx, currentHist.HistoryID, old, new, diff, now)
}

// CloseHistory closes history records for a deleted image.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	// Get current image history
	currentHist, err := tx.BronzeHistoryGCPComputeImage.Query().
		Where(
			bronzehistorygcpcomputeimage.ResourceID(resourceID),
			bronzehistorygcpcomputeimage.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("failed to find current image history: %w", err)
	}

	// Close image history
	if err := tx.BronzeHistoryGCPComputeImage.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("failed to close image history: %w", err)
	}

	// Close all children history
	return h.closeChildrenHistory(ctx, tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(ctx context.Context, tx *ent.Tx, imageHistoryID uint, data *ImageData, now time.Time) error {
	// Labels
	for _, labelData := range data.Labels {
		_, err := tx.BronzeHistoryGCPComputeImageLabel.Create().
			SetImageHistoryID(imageHistoryID).
			SetValidFrom(now).
			SetKey(labelData.Key).
			SetValue(labelData.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label history: %w", err)
		}
	}

	// Licenses
	for _, licenseData := range data.Licenses {
		_, err := tx.BronzeHistoryGCPComputeImageLicense.Create().
			SetImageHistoryID(imageHistoryID).
			SetValidFrom(now).
			SetLicense(licenseData.License).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create license history: %w", err)
		}
	}

	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(ctx context.Context, tx *ent.Tx, imageHistoryID uint, now time.Time) error {
	// Close labels
	_, err := tx.BronzeHistoryGCPComputeImageLabel.Update().
		Where(
			bronzehistorygcpcomputeimagelabel.ImageHistoryID(imageHistoryID),
			bronzehistorygcpcomputeimagelabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close label history: %w", err)
	}

	// Close licenses
	_, err = tx.BronzeHistoryGCPComputeImageLicense.Update().
		Where(
			bronzehistorygcpcomputeimagelicense.ImageHistoryID(imageHistoryID),
			bronzehistorygcpcomputeimagelicense.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close license history: %w", err)
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(ctx context.Context, tx *ent.Tx, imageHistoryID uint, old *ent.BronzeGCPComputeImage, new *ImageData, diff *ImageDiff, now time.Time) error {
	if diff.LabelsDiff.Changed {
		if err := h.updateLabelsHistory(ctx, tx, imageHistoryID, new.Labels, now); err != nil {
			return err
		}
	}

	if diff.LicensesDiff.Changed {
		if err := h.updateLicensesHistory(ctx, tx, imageHistoryID, new.Licenses, now); err != nil {
			return err
		}
	}

	return nil
}

func (h *HistoryService) updateLabelsHistory(ctx context.Context, tx *ent.Tx, imageHistoryID uint, labels []ImageLabelData, now time.Time) error {
	// Close old labels
	_, err := tx.BronzeHistoryGCPComputeImageLabel.Update().
		Where(
			bronzehistorygcpcomputeimagelabel.ImageHistoryID(imageHistoryID),
			bronzehistorygcpcomputeimagelabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close label history: %w", err)
	}

	// Create new labels
	for _, labelData := range labels {
		_, err := tx.BronzeHistoryGCPComputeImageLabel.Create().
			SetImageHistoryID(imageHistoryID).
			SetValidFrom(now).
			SetKey(labelData.Key).
			SetValue(labelData.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label history: %w", err)
		}
	}
	return nil
}

func (h *HistoryService) updateLicensesHistory(ctx context.Context, tx *ent.Tx, imageHistoryID uint, licenses []ImageLicenseData, now time.Time) error {
	// Close old licenses
	_, err := tx.BronzeHistoryGCPComputeImageLicense.Update().
		Where(
			bronzehistorygcpcomputeimagelicense.ImageHistoryID(imageHistoryID),
			bronzehistorygcpcomputeimagelicense.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close license history: %w", err)
	}

	// Create new licenses
	for _, licenseData := range licenses {
		_, err := tx.BronzeHistoryGCPComputeImageLicense.Create().
			SetImageHistoryID(imageHistoryID).
			SetValidFrom(now).
			SetLicense(licenseData.License).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create license history: %w", err)
		}
	}
	return nil
}
