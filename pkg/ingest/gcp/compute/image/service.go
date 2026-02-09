package image

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcomputeimage"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcomputeimagelabel"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcomputeimagelicense"
)

// Service handles GCP Compute image ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new image ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for image ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of image ingestion.
type IngestResult struct {
	ProjectID      string
	ImageCount     int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches images from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch images from GCP
	images, err := s.client.ListImages(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	// Convert to data structs
	imageDataList := make([]*ImageData, 0, len(images))
	for _, img := range images {
		data, err := ConvertImage(img, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert image: %w", err)
		}
		imageDataList = append(imageDataList, data)
	}

	// Save to database
	if err := s.saveImages(ctx, imageDataList); err != nil {
		return nil, fmt.Errorf("failed to save images: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		ImageCount:     len(imageDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveImages saves images to the database with history tracking.
func (s *Service) saveImages(ctx context.Context, images []*ImageData) error {
	if len(images) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, imageData := range images {
		// Load existing image with all relations
		existing, err := tx.BronzeGCPComputeImage.Query().
			Where(bronzegcpcomputeimage.ID(imageData.ID)).
			WithLabels().
			WithLicenses().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing image %s: %w", imageData.Name, err)
		}

		// Compute diff
		diff := DiffImageData(existing, imageData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPComputeImage.UpdateOneID(imageData.ID).
				SetCollectedAt(imageData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for image %s: %w", imageData.Name, err)
			}
			continue
		}

		// Delete old child entities if updating
		if existing != nil {
			if err := s.deleteImageChildren(ctx, tx, imageData.ID); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old children for image %s: %w", imageData.Name, err)
			}
		}

		// Create or update image
		var savedImage *ent.BronzeGCPComputeImage
		if existing == nil {
			// Create new image
			create := tx.BronzeGCPComputeImage.Create().
				SetID(imageData.ID).
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
				SetProjectID(imageData.ProjectID).
				SetCollectedAt(imageData.CollectedAt).
				SetFirstCollectedAt(imageData.CollectedAt)

			if imageData.ImageEncryptionKeyJSON != nil {
				create.SetImageEncryptionKeyJSON(imageData.ImageEncryptionKeyJSON)
			}
			if imageData.SourceDiskEncryptionKeyJSON != nil {
				create.SetSourceDiskEncryptionKeyJSON(imageData.SourceDiskEncryptionKeyJSON)
			}
			if imageData.SourceImageEncryptionKeyJSON != nil {
				create.SetSourceImageEncryptionKeyJSON(imageData.SourceImageEncryptionKeyJSON)
			}
			if imageData.SourceSnapshotEncryptionKeyJSON != nil {
				create.SetSourceSnapshotEncryptionKeyJSON(imageData.SourceSnapshotEncryptionKeyJSON)
			}
			if imageData.DeprecatedJSON != nil {
				create.SetDeprecatedJSON(imageData.DeprecatedJSON)
			}
			if imageData.GuestOsFeaturesJSON != nil {
				create.SetGuestOsFeaturesJSON(imageData.GuestOsFeaturesJSON)
			}
			if imageData.ShieldedInstanceInitialStateJSON != nil {
				create.SetShieldedInstanceInitialStateJSON(imageData.ShieldedInstanceInitialStateJSON)
			}
			if imageData.RawDiskJSON != nil {
				create.SetRawDiskJSON(imageData.RawDiskJSON)
			}
			if imageData.StorageLocationsJSON != nil {
				create.SetStorageLocationsJSON(imageData.StorageLocationsJSON)
			}
			if imageData.LicenseCodesJSON != nil {
				create.SetLicenseCodesJSON(imageData.LicenseCodesJSON)
			}

			savedImage, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create image %s: %w", imageData.Name, err)
			}
		} else {
			// Update existing image
			update := tx.BronzeGCPComputeImage.UpdateOneID(imageData.ID).
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
				SetProjectID(imageData.ProjectID).
				SetCollectedAt(imageData.CollectedAt)

			if imageData.ImageEncryptionKeyJSON != nil {
				update.SetImageEncryptionKeyJSON(imageData.ImageEncryptionKeyJSON)
			}
			if imageData.SourceDiskEncryptionKeyJSON != nil {
				update.SetSourceDiskEncryptionKeyJSON(imageData.SourceDiskEncryptionKeyJSON)
			}
			if imageData.SourceImageEncryptionKeyJSON != nil {
				update.SetSourceImageEncryptionKeyJSON(imageData.SourceImageEncryptionKeyJSON)
			}
			if imageData.SourceSnapshotEncryptionKeyJSON != nil {
				update.SetSourceSnapshotEncryptionKeyJSON(imageData.SourceSnapshotEncryptionKeyJSON)
			}
			if imageData.DeprecatedJSON != nil {
				update.SetDeprecatedJSON(imageData.DeprecatedJSON)
			}
			if imageData.GuestOsFeaturesJSON != nil {
				update.SetGuestOsFeaturesJSON(imageData.GuestOsFeaturesJSON)
			}
			if imageData.ShieldedInstanceInitialStateJSON != nil {
				update.SetShieldedInstanceInitialStateJSON(imageData.ShieldedInstanceInitialStateJSON)
			}
			if imageData.RawDiskJSON != nil {
				update.SetRawDiskJSON(imageData.RawDiskJSON)
			}
			if imageData.StorageLocationsJSON != nil {
				update.SetStorageLocationsJSON(imageData.StorageLocationsJSON)
			}
			if imageData.LicenseCodesJSON != nil {
				update.SetLicenseCodesJSON(imageData.LicenseCodesJSON)
			}

			savedImage, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update image %s: %w", imageData.Name, err)
			}
		}

		// Create child entities
		if err := s.createImageChildren(ctx, tx, savedImage, imageData); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create children for image %s: %w", imageData.Name, err)
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, imageData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for image %s: %w", imageData.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, imageData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for image %s: %w", imageData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// deleteImageChildren deletes all child entities for an image.
// Note: Ent CASCADE DELETE is set on edges, so we just need to delete direct children.
func (s *Service) deleteImageChildren(ctx context.Context, tx *ent.Tx, imageID string) error {
	// Delete labels
	_, err := tx.BronzeGCPComputeImageLabel.Delete().
		Where(bronzegcpcomputeimagelabel.HasImageWith(bronzegcpcomputeimage.ID(imageID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete labels: %w", err)
	}

	// Delete licenses
	_, err = tx.BronzeGCPComputeImageLicense.Delete().
		Where(bronzegcpcomputeimagelicense.HasImageWith(bronzegcpcomputeimage.ID(imageID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete licenses: %w", err)
	}

	return nil
}

// createImageChildren creates all child entities for an image.
func (s *Service) createImageChildren(ctx context.Context, tx *ent.Tx, image *ent.BronzeGCPComputeImage, data *ImageData) error {
	// Create labels
	for _, labelData := range data.Labels {
		_, err := tx.BronzeGCPComputeImageLabel.Create().
			SetImage(image).
			SetKey(labelData.Key).
			SetValue(labelData.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label: %w", err)
		}
	}

	// Create licenses
	for _, licenseData := range data.Licenses {
		_, err := tx.BronzeGCPComputeImageLicense.Create().
			SetImage(image).
			SetLicense(licenseData.License).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create license: %w", err)
		}
	}

	return nil
}

// DeleteStaleImages removes images that were not collected in the latest run.
// Also closes history records for deleted images.
func (s *Service) DeleteStaleImages(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	// Find stale images
	staleImages, err := tx.BronzeGCPComputeImage.Query().
		Where(
			bronzegcpcomputeimage.ProjectID(projectID),
			bronzegcpcomputeimage.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale image
	for _, img := range staleImages {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, img.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for image %s: %w", img.ID, err)
		}

		// Delete image (CASCADE will delete children)
		if err := tx.BronzeGCPComputeImage.DeleteOne(img).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete image %s: %w", img.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
