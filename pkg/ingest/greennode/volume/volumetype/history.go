package volumetype

import (
	"context"
	"fmt"
	"time"

	entvol "danny.vn/hotpot/pkg/storage/ent/greennode/volume"
	"danny.vn/hotpot/pkg/storage/ent/greennode/volume/bronzehistorygreennodevolumevolumetype"
)

// HistoryService handles history tracking for volume types.
type HistoryService struct {
	entClient *entvol.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entvol.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new volume type.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entvol.Tx, data *VolumeTypeData, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeVolumeVolumeType.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetIops(data.Iops).
		SetMaxSize(data.MaxSize).
		SetMinSize(data.MinSize).
		SetThroughPut(data.ThroughPut).
		SetZoneID(data.ZoneID).
		SetRegion(data.Region).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create volume type history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entvol.Tx, old *entvol.BronzeGreenNodeVolumeVolumeType, new *VolumeTypeData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeVolumeVolumeType.Query().
		Where(
			bronzehistorygreennodevolumevolumetype.ResourceID(old.ID),
			bronzehistorygreennodevolumevolumetype.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current volume type history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeVolumeVolumeType.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close volume type history: %w", err)
	}

	_, err = tx.BronzeHistoryGreenNodeVolumeVolumeType.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetIops(new.Iops).
		SetMaxSize(new.MaxSize).
		SetMinSize(new.MinSize).
		SetThroughPut(new.ThroughPut).
		SetZoneID(new.ZoneID).
		SetRegion(new.Region).
		SetProjectID(new.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new volume type history: %w", err)
	}
	return nil
}

// CloseHistory closes history for a deleted volume type.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entvol.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeVolumeVolumeType.Query().
		Where(
			bronzehistorygreennodevolumevolumetype.ResourceID(resourceID),
			bronzehistorygreennodevolumevolumetype.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entvol.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current volume type history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeVolumeVolumeType.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close volume type history: %w", err)
	}
	return nil
}
