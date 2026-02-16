package volume

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorydovolume"
)

// HistoryService handles history tracking for Volumes.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) buildCreate(tx *ent.Tx, data *VolumeData) *ent.BronzeHistoryDOVolumeCreate {
	create := tx.BronzeHistoryDOVolume.Create().
		SetResourceID(data.ResourceID).
		SetName(data.Name).
		SetRegion(data.Region).
		SetSizeGigabytes(data.SizeGigabytes).
		SetDescription(data.Description).
		SetDropletIdsJSON(data.DropletIdsJSON).
		SetFilesystemType(data.FilesystemType).
		SetFilesystemLabel(data.FilesystemLabel).
		SetTagsJSON(data.TagsJSON)

	if data.APICreatedAt != nil {
		create.SetAPICreatedAt(*data.APICreatedAt)
	}

	return create
}

// CreateHistory creates a history record for a new Volume.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *VolumeData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create volume history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new for a changed Volume.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeDOVolume, new *VolumeData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDOVolume.Query().
		Where(
			bronzehistorydovolume.ResourceID(old.ID),
			bronzehistorydovolume.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current volume history: %w", err)
	}

	if err := tx.BronzeHistoryDOVolume.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close volume history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new volume history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted Volume.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDOVolume.Query().
		Where(
			bronzehistorydovolume.ResourceID(resourceID),
			bronzehistorydovolume.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current volume history: %w", err)
	}

	if err := tx.BronzeHistoryDOVolume.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close volume history: %w", err)
	}

	return nil
}
