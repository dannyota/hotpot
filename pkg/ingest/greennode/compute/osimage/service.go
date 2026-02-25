package osimage

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegreennodecomputeosimage"
)

// Service handles GreenNode OS image ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new OS image ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of OS image ingestion.
type IngestResult struct {
	OSImageCount   int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches OS images from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID, region string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	images, err := s.client.ListOSImages(ctx)
	if err != nil {
		return nil, fmt.Errorf("list os images: %w", err)
	}

	dataList := make([]*OSImageData, 0, len(images))
	for _, img := range images {
		dataList = append(dataList, ConvertOSImage(img, projectID, region, collectedAt))
	}

	if err := s.saveOSImages(ctx, dataList); err != nil {
		return nil, fmt.Errorf("save os images: %w", err)
	}

	return &IngestResult{
		OSImageCount:   len(dataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveOSImages(ctx context.Context, images []*OSImageData) error {
	if len(images) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, data := range images {
		existing, err := tx.BronzeGreenNodeComputeOSImage.Query().
			Where(bronzegreennodecomputeosimage.ID(data.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing os image %s: %w", data.ID, err)
		}

		diff := DiffOSImageData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodeComputeOSImage.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for os image %s: %w", data.ID, err)
			}
			continue
		}

		if existing == nil {
			_, err = tx.BronzeGreenNodeComputeOSImage.Create().
				SetID(data.ID).
				SetImageType(data.ImageType).
				SetImageVersion(data.ImageVersion).
				SetNillableLicence(data.Licence).
				SetNillableLicenseKey(data.LicenseKey).
				SetDescription(data.Description).
				SetZoneID(data.ZoneID).
				SetFlavorZoneIds(data.FlavorZoneIDs).
				SetDefaultTagIds(data.DefaultTagIDs).
				SetPackageLimitCPU(data.PackageLimitCpu).
				SetPackageLimitMemory(data.PackageLimitMemory).
				SetPackageLimitDiskSize(data.PackageLimitDiskSize).
				SetRegion(data.Region).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create os image %s: %w", data.ID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for os image %s: %w", data.ID, err)
			}
		} else {
			_, err = tx.BronzeGreenNodeComputeOSImage.UpdateOneID(data.ID).
				SetImageType(data.ImageType).
				SetImageVersion(data.ImageVersion).
				SetNillableLicence(data.Licence).
				SetNillableLicenseKey(data.LicenseKey).
				SetDescription(data.Description).
				SetZoneID(data.ZoneID).
				SetFlavorZoneIds(data.FlavorZoneIDs).
				SetDefaultTagIds(data.DefaultTagIDs).
				SetPackageLimitCPU(data.PackageLimitCpu).
				SetPackageLimitMemory(data.PackageLimitMemory).
				SetPackageLimitDiskSize(data.PackageLimitDiskSize).
				SetRegion(data.Region).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update os image %s: %w", data.ID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for os image %s: %w", data.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// DeleteStaleOSImages removes OS images not collected in the latest run for the given region.
func (s *Service) DeleteStaleOSImages(ctx context.Context, projectID, region string, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	stale, err := tx.BronzeGreenNodeComputeOSImage.Query().
		Where(
			bronzegreennodecomputeosimage.ProjectID(projectID),
			bronzegreennodecomputeosimage.Region(region),
			bronzegreennodecomputeosimage.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale os images: %w", err)
	}

	for _, img := range stale {
		if err := s.history.CloseHistory(ctx, tx, img.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for os image %s: %w", img.ID, err)
		}
		if err := tx.BronzeGreenNodeComputeOSImage.DeleteOneID(img.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete os image %s: %w", img.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
