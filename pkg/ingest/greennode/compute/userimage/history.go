package userimage

import (
	"context"
	"fmt"
	"time"

	entcompute "github.com/dannyota/hotpot/pkg/storage/ent/greennode/compute"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/compute/bronzehistorygreennodecomputeuserimage"
)

// HistoryService handles history tracking for user images.
type HistoryService struct {
	entClient *entcompute.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entcompute.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new user image.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entcompute.Tx, data *UserImageData, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeComputeUserImage.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetStatus(data.Status).
		SetMinDisk(data.MinDisk).
		SetImageSize(data.ImageSize).
		SetMetaData(data.MetaData).
		SetCreatedAt(data.CreatedAt).
		SetRegion(data.Region).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create user image history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entcompute.Tx, old *entcompute.BronzeGreenNodeComputeUserImage, new *UserImageData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeComputeUserImage.Query().
		Where(
			bronzehistorygreennodecomputeuserimage.ResourceID(old.ID),
			bronzehistorygreennodecomputeuserimage.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current user image history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeComputeUserImage.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close user image history: %w", err)
	}

	_, err = tx.BronzeHistoryGreenNodeComputeUserImage.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetStatus(new.Status).
		SetMinDisk(new.MinDisk).
		SetImageSize(new.ImageSize).
		SetMetaData(new.MetaData).
		SetCreatedAt(new.CreatedAt).
		SetRegion(new.Region).
		SetProjectID(new.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new user image history: %w", err)
	}
	return nil
}

// CloseHistory closes history for a deleted user image.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entcompute.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeComputeUserImage.Query().
		Where(
			bronzehistorygreennodecomputeuserimage.ResourceID(resourceID),
			bronzehistorygreennodecomputeuserimage.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entcompute.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current user image history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeComputeUserImage.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close user image history: %w", err)
	}
	return nil
}
