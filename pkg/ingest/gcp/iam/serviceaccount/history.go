package serviceaccount

import (
	"context"
	"fmt"
	"time"

	entiam "github.com/dannyota/hotpot/pkg/storage/ent/gcp/iam"
	"github.com/dannyota/hotpot/pkg/storage/ent/gcp/iam/bronzehistorygcpiamserviceaccount"
)

type HistoryService struct {
	entClient *entiam.Client
}

func NewHistoryService(entClient *entiam.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) CreateHistory(ctx context.Context, tx *entiam.Tx, saData *ServiceAccountData, now time.Time) error {
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

func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entiam.Tx, old *entiam.BronzeGCPIAMServiceAccount, new *ServiceAccountData, diff *ServiceAccountDiff, now time.Time) error {
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

func (h *HistoryService) CloseHistory(ctx context.Context, tx *entiam.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPIAMServiceAccount.Update().
		Where(
			bronzehistorygcpiamserviceaccount.ResourceID(resourceID),
			bronzehistorygcpiamserviceaccount.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if entiam.IsNotFound(err) {
		return nil // No history to close
	}
	return err
}
