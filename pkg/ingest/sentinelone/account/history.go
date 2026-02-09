package account

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorys1account"
)

// HistoryService handles history tracking for accounts.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new account.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *AccountData, now time.Time) error {
	_, err := tx.BronzeHistoryS1Account.Create().
		SetResourceID(data.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create account history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history for a changed account.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeS1Account, new *AccountData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1Account.Query().
		Where(
			bronzehistorys1account.ResourceID(old.ID),
			bronzehistorys1account.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current account history: %w", err)
	}

	if err := tx.BronzeHistoryS1Account.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close account history: %w", err)
	}

	_, err = tx.BronzeHistoryS1Account.Create().
		SetResourceID(new.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new account history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted account.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1Account.Query().
		Where(
			bronzehistorys1account.ResourceID(resourceID),
			bronzehistorys1account.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current account history: %w", err)
	}

	if err := tx.BronzeHistoryS1Account.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close account history: %w", err)
	}

	return nil
}
