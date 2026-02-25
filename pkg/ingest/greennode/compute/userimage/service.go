package userimage

import (
	"context"
	"fmt"
	"time"

	entcompute "github.com/dannyota/hotpot/pkg/storage/ent/greennode/compute"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/compute/bronzegreennodecomputeuserimage"
)

// Service handles GreenNode user image ingestion.
type Service struct {
	client    *Client
	entClient *entcompute.Client
	history   *HistoryService
}

// NewService creates a new user image ingestion service.
func NewService(client *Client, entClient *entcompute.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of user image ingestion.
type IngestResult struct {
	UserImageCount int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches user images from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID, region string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	images, err := s.client.ListUserImages(ctx)
	if err != nil {
		return nil, fmt.Errorf("list user images: %w", err)
	}

	dataList := make([]*UserImageData, 0, len(images))
	for _, img := range images {
		dataList = append(dataList, ConvertUserImage(img, projectID, region, collectedAt))
	}

	if err := s.saveUserImages(ctx, dataList); err != nil {
		return nil, fmt.Errorf("save user images: %w", err)
	}

	return &IngestResult{
		UserImageCount: len(dataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveUserImages(ctx context.Context, images []*UserImageData) error {
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
		existing, err := tx.BronzeGreenNodeComputeUserImage.Query().
			Where(bronzegreennodecomputeuserimage.ID(data.ID)).
			First(ctx)
		if err != nil && !entcompute.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing user image %s: %w", data.Name, err)
		}

		diff := DiffUserImageData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodeComputeUserImage.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for user image %s: %w", data.Name, err)
			}
			continue
		}

		if existing == nil {
			_, err = tx.BronzeGreenNodeComputeUserImage.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetStatus(data.Status).
				SetMinDisk(data.MinDisk).
				SetImageSize(data.ImageSize).
				SetMetaData(data.MetaData).
				SetCreatedAt(data.CreatedAt).
				SetRegion(data.Region).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create user image %s: %w", data.Name, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for user image %s: %w", data.Name, err)
			}
		} else {
			_, err = tx.BronzeGreenNodeComputeUserImage.UpdateOneID(data.ID).
				SetName(data.Name).
				SetStatus(data.Status).
				SetMinDisk(data.MinDisk).
				SetImageSize(data.ImageSize).
				SetMetaData(data.MetaData).
				SetCreatedAt(data.CreatedAt).
				SetRegion(data.Region).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update user image %s: %w", data.Name, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for user image %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// DeleteStaleUserImages removes user images not collected in the latest run for the given region.
func (s *Service) DeleteStaleUserImages(ctx context.Context, projectID, region string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGreenNodeComputeUserImage.Query().
		Where(
			bronzegreennodecomputeuserimage.ProjectID(projectID),
			bronzegreennodecomputeuserimage.Region(region),
			bronzegreennodecomputeuserimage.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale user images: %w", err)
	}

	for _, img := range stale {
		if err := s.history.CloseHistory(ctx, tx, img.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for user image %s: %w", img.ID, err)
		}
		if err := tx.BronzeGreenNodeComputeUserImage.DeleteOneID(img.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete user image %s: %w", img.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
