package glbregion

import (
	"context"
	"fmt"
	"time"

	entglb "github.com/dannyota/hotpot/pkg/storage/ent/greennode/glb"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/glb/bronzehistorygreennodeglbglobalregion"
)

// HistoryService handles history tracking for global regions.
type HistoryService struct {
	entClient *entglb.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entglb.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new global region.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entglb.Tx, data *GLBRegionData, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeGLBGlobalRegion.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetVserverEndpoint(data.VserverEndpoint).
		SetVlbEndpoint(data.VlbEndpoint).
		SetUIServerEndpoint(data.UIServerEndpoint).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create region history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entglb.Tx, old *entglb.BronzeGreenNodeGLBGlobalRegion, new *GLBRegionData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeGLBGlobalRegion.Query().
		Where(
			bronzehistorygreennodeglbglobalregion.ResourceID(old.ID),
			bronzehistorygreennodeglbglobalregion.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current region history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeGLBGlobalRegion.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close region history: %w", err)
	}

	_, err = tx.BronzeHistoryGreenNodeGLBGlobalRegion.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetVserverEndpoint(new.VserverEndpoint).
		SetVlbEndpoint(new.VlbEndpoint).
		SetUIServerEndpoint(new.UIServerEndpoint).
		SetProjectID(new.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new region history: %w", err)
	}
	return nil
}

// CloseHistory closes history for a deleted region.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entglb.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeGLBGlobalRegion.Query().
		Where(
			bronzehistorygreennodeglbglobalregion.ResourceID(resourceID),
			bronzehistorygreennodeglbglobalregion.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entglb.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current region history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeGLBGlobalRegion.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close region history: %w", err)
	}
	return nil
}
