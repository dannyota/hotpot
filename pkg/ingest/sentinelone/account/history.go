package account

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorys1account"
)

// HistoryService handles history tracking for accounts.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) buildCreate(tx *ent.Tx, data *AccountData) *ent.BronzeHistoryS1AccountCreate {
	create := tx.BronzeHistoryS1Account.Create().
		SetResourceID(data.ResourceID).
		SetName(data.Name).
		SetState(data.State).
		SetAccountType(data.AccountType).
		SetUnlimitedExpiration(data.UnlimitedExpiration).
		SetActiveAgents(data.ActiveAgents).
		SetTotalLicenses(data.TotalLicenses).
		SetUsageType(data.UsageType).
		SetBillingMode(data.BillingMode).
		SetCreator(data.Creator).
		SetCreatorID(data.CreatorID).
		SetNumberOfSites(data.NumberOfSites).
		SetExternalID(data.ExternalID)

	if data.APICreatedAt != nil {
		create.SetAPICreatedAt(*data.APICreatedAt)
	}
	if data.APIUpdatedAt != nil {
		create.SetAPIUpdatedAt(*data.APIUpdatedAt)
	}
	if data.Expiration != nil {
		create.SetExpiration(*data.Expiration)
	}
	if data.LicensesJSON != nil {
		create.SetLicensesJSON(data.LicensesJSON)
	}

	return create
}

// CreateHistory creates a history record for a new account.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *AccountData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
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

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
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
