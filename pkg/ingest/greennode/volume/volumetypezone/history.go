package volumetypezone

import (
	"context"
	"fmt"
	"time"

	entvol "github.com/dannyota/hotpot/pkg/storage/ent/greennode/volume"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/volume/bronzehistorygreennodevolumevolumetypezone"
)

// HistoryService handles history tracking for volume type zones.
type HistoryService struct {
	entClient *entvol.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entvol.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new volume type zone.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entvol.Tx, data *VolumeTypeZoneData, now time.Time) error {
	create := tx.BronzeHistoryGreenNodeVolumeVolumeTypeZone.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetRegion(data.Region).
		SetProjectID(data.ProjectID)

	if data.PoolNameJSON != nil {
		create.SetPoolNameJSON(data.PoolNameJSON)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("create volume type zone history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entvol.Tx, old *entvol.BronzeGreenNodeVolumeVolumeTypeZone, new *VolumeTypeZoneData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeVolumeVolumeTypeZone.Query().
		Where(
			bronzehistorygreennodevolumevolumetypezone.ResourceID(old.ID),
			bronzehistorygreennodevolumevolumetypezone.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current volume type zone history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeVolumeVolumeTypeZone.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close volume type zone history: %w", err)
	}

	create := tx.BronzeHistoryGreenNodeVolumeVolumeTypeZone.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetRegion(new.Region).
		SetProjectID(new.ProjectID)

	if new.PoolNameJSON != nil {
		create.SetPoolNameJSON(new.PoolNameJSON)
	}

	_, err = create.Save(ctx)
	if err != nil {
		return fmt.Errorf("create new volume type zone history: %w", err)
	}
	return nil
}

// CloseHistory closes history for a deleted volume type zone.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entvol.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeVolumeVolumeTypeZone.Query().
		Where(
			bronzehistorygreennodevolumevolumetypezone.ResourceID(resourceID),
			bronzehistorygreennodevolumevolumetypezone.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entvol.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current volume type zone history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeVolumeVolumeTypeZone.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close volume type zone history: %w", err)
	}
	return nil
}
