package app_inventory

import (
	"context"
	"fmt"
	"time"

	ents1 "github.com/dannyota/hotpot/pkg/storage/ent/s1"
	"github.com/dannyota/hotpot/pkg/storage/ent/s1/bronzehistorys1appinventory"
)

// HistoryService handles history tracking for app inventory.
type HistoryService struct {
	entClient *ents1.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ents1.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) buildCreate(tx *ents1.Tx, data *AppInventoryData) *ents1.BronzeHistoryS1AppInventoryCreate {
	return tx.BronzeHistoryS1AppInventory.Create().
		SetResourceID(data.ResourceID).
		SetApplicationName(data.ApplicationName).
		SetApplicationVendor(data.ApplicationVendor).
		SetEndpointsCount(data.EndpointsCount).
		SetApplicationVersionsCount(data.ApplicationVersionsCount).
		SetEstimate(data.Estimate)
}

// CreateHistory creates a history record for a new app inventory entry.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ents1.Tx, data *AppInventoryData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create app inventory history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history for a changed app inventory entry.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ents1.Tx, old *ents1.BronzeS1AppInventory, new *AppInventoryData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1AppInventory.Query().
		Where(
			bronzehistorys1appinventory.ResourceID(old.ID),
			bronzehistorys1appinventory.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current app inventory history: %w", err)
	}

	if err := tx.BronzeHistoryS1AppInventory.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close app inventory history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new app inventory history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted app inventory entry.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ents1.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1AppInventory.Query().
		Where(
			bronzehistorys1appinventory.ResourceID(resourceID),
			bronzehistorys1appinventory.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ents1.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current app inventory history: %w", err)
	}

	if err := tx.BronzeHistoryS1AppInventory.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close app inventory history: %w", err)
	}

	return nil
}
