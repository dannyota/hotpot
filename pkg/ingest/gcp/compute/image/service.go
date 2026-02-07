package image

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
)

// Service handles GCP Compute image ingestion.
type Service struct {
	client  *Client
	db      *gorm.DB
	history *HistoryService
}

// NewService creates a new image ingestion service.
func NewService(client *Client, db *gorm.DB) *Service {
	return &Service{
		client:  client,
		db:      db,
		history: NewHistoryService(db),
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

	// Convert to bronze models
	bronzeImages := make([]bronze.GCPComputeImage, 0, len(images))
	for _, img := range images {
		bi, err := ConvertImage(img, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert image: %w", err)
		}
		bronzeImages = append(bronzeImages, bi)
	}

	// Save to database
	if err := s.saveImages(ctx, bronzeImages); err != nil {
		return nil, fmt.Errorf("failed to save images: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		ImageCount:     len(bronzeImages),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveImages saves images to the database with history tracking.
func (s *Service) saveImages(ctx context.Context, images []bronze.GCPComputeImage) error {
	if len(images) == 0 {
		return nil
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, img := range images {
			// Load existing image with all relations
			var existing *bronze.GCPComputeImage
			var old bronze.GCPComputeImage
			err := tx.Preload("Labels").Preload("Licenses").
				Where("resource_id = ?", img.ResourceID).
				First(&old).Error
			if err == nil {
				existing = &old
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to load existing image %s: %w", img.Name, err)
			}

			// Compute diff
			diff := DiffImage(existing, &img)

			// Skip if no changes
			if !diff.HasAnyChange() && existing != nil {
				// Update collected_at only
				if err := tx.Model(&bronze.GCPComputeImage{}).
					Where("resource_id = ?", img.ResourceID).
					Update("collected_at", img.CollectedAt).Error; err != nil {
					return fmt.Errorf("failed to update collected_at for image %s: %w", img.Name, err)
				}
				continue
			}

			// Delete old relations (manual cascade)
			if existing != nil {
				if err := s.deleteImageRelations(tx, img.ResourceID); err != nil {
					return fmt.Errorf("failed to delete old relations for image %s: %w", img.Name, err)
				}
			}

			// Upsert image
			if err := tx.Save(&img).Error; err != nil {
				return fmt.Errorf("failed to upsert image %s: %w", img.Name, err)
			}

			// Create new relations
			if err := s.createImageRelations(tx, img.ResourceID, &img); err != nil {
				return fmt.Errorf("failed to create relations for image %s: %w", img.Name, err)
			}

			// Track history
			if diff.IsNew {
				if err := s.history.CreateHistory(tx, &img, now); err != nil {
					return fmt.Errorf("failed to create history for image %s: %w", img.Name, err)
				}
			} else {
				if err := s.history.UpdateHistory(tx, existing, &img, diff, now); err != nil {
					return fmt.Errorf("failed to update history for image %s: %w", img.Name, err)
				}
			}
		}

		return nil
	})
}

// deleteImageRelations deletes all related records for an image.
func (s *Service) deleteImageRelations(tx *gorm.DB, imageResourceID string) error {
	// Delete labels
	if err := tx.Where("image_resource_id = ?", imageResourceID).Delete(&bronze.GCPComputeImageLabel{}).Error; err != nil {
		return err
	}

	// Delete licenses
	if err := tx.Where("image_resource_id = ?", imageResourceID).Delete(&bronze.GCPComputeImageLicense{}).Error; err != nil {
		return err
	}

	return nil
}

// createImageRelations creates all related records for an image.
func (s *Service) createImageRelations(tx *gorm.DB, imageResourceID string, img *bronze.GCPComputeImage) error {
	// Create labels
	for i := range img.Labels {
		img.Labels[i].ImageResourceID = imageResourceID
	}
	if len(img.Labels) > 0 {
		if err := tx.Create(&img.Labels).Error; err != nil {
			return fmt.Errorf("failed to create labels: %w", err)
		}
	}

	// Create licenses
	for i := range img.Licenses {
		img.Licenses[i].ImageResourceID = imageResourceID
	}
	if len(img.Licenses) > 0 {
		if err := tx.Create(&img.Licenses).Error; err != nil {
			return fmt.Errorf("failed to create licenses: %w", err)
		}
	}

	return nil
}

// DeleteStaleImages removes images that were not collected in the latest run.
// Also closes history records for deleted images.
func (s *Service) DeleteStaleImages(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find stale images
		var staleImages []bronze.GCPComputeImage
		if err := tx.Where("project_id = ? AND collected_at < ?", projectID, collectedAt).
			Find(&staleImages).Error; err != nil {
			return err
		}

		// Close history and delete each stale image
		for _, img := range staleImages {
			// Close history
			if err := s.history.CloseHistory(tx, img.ResourceID, now); err != nil {
				return fmt.Errorf("failed to close history for image %s: %w", img.ResourceID, err)
			}

			// Delete relations
			if err := s.deleteImageRelations(tx, img.ResourceID); err != nil {
				return fmt.Errorf("failed to delete relations for image %s: %w", img.ResourceID, err)
			}

			// Delete image
			if err := tx.Delete(&img).Error; err != nil {
				return fmt.Errorf("failed to delete image %s: %w", img.ResourceID, err)
			}
		}

		return nil
	})
}
