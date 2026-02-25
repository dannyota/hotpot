package zone

import (
	"context"
	"fmt"
	"time"

	entportal "github.com/dannyota/hotpot/pkg/storage/ent/greennode/portal"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/portal/bronzehistorygreennodeportalzone"
)

// HistoryService handles history tracking for zones.
type HistoryService struct {
	entClient *entportal.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entportal.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new zone.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entportal.Tx, data *ZoneData, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodePortalZone.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetOpenstackZone(data.OpenstackZone).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create zone history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entportal.Tx, old *entportal.BronzeGreenNodePortalZone, new *ZoneData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodePortalZone.Query().
		Where(
			bronzehistorygreennodeportalzone.ResourceID(old.ID),
			bronzehistorygreennodeportalzone.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current zone history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodePortalZone.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close zone history: %w", err)
	}

	_, err = tx.BronzeHistoryGreenNodePortalZone.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetOpenstackZone(new.OpenstackZone).
		SetProjectID(new.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new zone history: %w", err)
	}
	return nil
}

// CloseHistory closes history for a deleted zone.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entportal.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodePortalZone.Query().
		Where(
			bronzehistorygreennodeportalzone.ResourceID(resourceID),
			bronzehistorygreennodeportalzone.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entportal.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current zone history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodePortalZone.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close zone history: %w", err)
	}
	return nil
}
