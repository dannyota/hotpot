package account

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorydoaccount"
)

// HistoryService handles history tracking for Accounts.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) buildCreate(tx *ent.Tx, data *AccountData) *ent.BronzeHistoryDOAccountCreate {
	return tx.BronzeHistoryDOAccount.Create().
		SetResourceID(data.ResourceID).
		SetEmail(data.Email).
		SetName(data.Name).
		SetStatus(data.Status).
		SetStatusMessage(data.StatusMessage).
		SetDropletLimit(data.DropletLimit).
		SetFloatingIPLimit(data.FloatingIPLimit).
		SetReservedIPLimit(data.ReservedIPLimit).
		SetVolumeLimit(data.VolumeLimit).
		SetEmailVerified(data.EmailVerified).
		SetTeamName(data.TeamName).
		SetTeamUUID(data.TeamUUID)
}

// CreateHistory creates a history record for a new Account.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *AccountData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create Account history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new for a changed Account.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeDOAccount, new *AccountData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDOAccount.Query().
		Where(
			bronzehistorydoaccount.ResourceID(old.ID),
			bronzehistorydoaccount.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current Account history: %w", err)
	}

	if err := tx.BronzeHistoryDOAccount.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close Account history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new Account history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted Account.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDOAccount.Query().
		Where(
			bronzehistorydoaccount.ResourceID(resourceID),
			bronzehistorydoaccount.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current Account history: %w", err)
	}

	if err := tx.BronzeHistoryDOAccount.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close Account history: %w", err)
	}

	return nil
}
