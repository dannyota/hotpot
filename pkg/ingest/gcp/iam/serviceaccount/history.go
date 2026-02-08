package serviceaccount

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorygcpiamserviceaccount"
)

type HistoryService struct {
	entClient *ent.Client
}

func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, saData *ServiceAccountData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPIAMServiceAccount.Create().
		SetResourceID(saData.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(saData.CollectedAt).
		SetFirstCollectedAt(saData.CollectedAt).
		SetName(saData.Name).
		SetEmail(saData.Email).
		SetDisplayName(saData.DisplayName).
		SetDescription(saData.Description).
		SetOauth2ClientID(saData.Oauth2ClientId).
		SetDisabled(saData.Disabled).
		SetEtag(saData.Etag).
		SetProjectID(saData.ProjectID).
		Save(ctx)
	return err
}

func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPIAMServiceAccount, new *ServiceAccountData, diff *ServiceAccountDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	_, err := tx.BronzeHistoryGCPIAMServiceAccount.Update().
		Where(
			bronzehistorygcpiamserviceaccount.ResourceID(old.ID),
			bronzehistorygcpiamserviceaccount.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close old history: %w", err)
	}

	// Create new history
	_, err = tx.BronzeHistoryGCPIAMServiceAccount.Create().
		SetResourceID(new.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetEmail(new.Email).
		SetDisplayName(new.DisplayName).
		SetDescription(new.Description).
		SetOauth2ClientID(new.Oauth2ClientId).
		SetDisabled(new.Disabled).
		SetEtag(new.Etag).
		SetProjectID(new.ProjectID).
		Save(ctx)
	return err
}

func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPIAMServiceAccount.Update().
		Where(
			bronzehistorygcpiamserviceaccount.ResourceID(resourceID),
			bronzehistorygcpiamserviceaccount.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if ent.IsNotFound(err) {
		return nil // No history to close
	}
	return err
}
