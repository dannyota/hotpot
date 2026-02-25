package region

import (
	"context"
	"fmt"
	"time"

	entportal "github.com/dannyota/hotpot/pkg/storage/ent/greennode/portal"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/portal/bronzehistorygreennodeportalregion"
)

// HistoryService handles history tracking for regions.
type HistoryService struct {
	entClient *entportal.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entportal.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new region.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entportal.Tx, data *RegionData, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodePortalRegion.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetDescription(data.Description).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create region history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entportal.Tx, old *entportal.BronzeGreenNodePortalRegion, new *RegionData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodePortalRegion.Query().
		Where(
			bronzehistorygreennodeportalregion.ResourceID(old.ID),
			bronzehistorygreennodeportalregion.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current region history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodePortalRegion.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close region history: %w", err)
	}

	_, err = tx.BronzeHistoryGreenNodePortalRegion.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetDescription(new.Description).
		SetProjectID(new.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new region history: %w", err)
	}
	return nil
}

// CloseHistory closes history for a deleted region.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entportal.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodePortalRegion.Query().
		Where(
			bronzehistorygreennodeportalregion.ResourceID(resourceID),
			bronzehistorygreennodeportalregion.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entportal.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current region history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodePortalRegion.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close region history: %w", err)
	}
	return nil
}
