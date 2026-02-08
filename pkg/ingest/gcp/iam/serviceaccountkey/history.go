package serviceaccountkey

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorygcpiamserviceaccountkey"
)

// HistoryService manages service account key history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new service account key.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *ServiceAccountKeyData, now time.Time) error {
	create := tx.BronzeHistoryGCPIAMServiceAccountKey.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetServiceAccountEmail(data.ServiceAccountEmail).
		SetDisabled(data.Disabled).
		SetProjectID(data.ProjectID)

	if data.KeyOrigin != "" {
		create.SetKeyOrigin(data.KeyOrigin)
	}
	if data.KeyType != "" {
		create.SetKeyType(data.KeyType)
	}
	if data.KeyAlgorithm != "" {
		create.SetKeyAlgorithm(data.KeyAlgorithm)
	}
	if !data.ValidAfterTime.IsZero() {
		create.SetValidAfterTime(data.ValidAfterTime)
	}
	if !data.ValidBeforeTime.IsZero() {
		create.SetValidBeforeTime(data.ValidBeforeTime)
	}

	_, err := create.Save(ctx)
	return err
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPIAMServiceAccountKey, new *ServiceAccountKeyData, diff *ServiceAccountKeyDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	_, err := tx.BronzeHistoryGCPIAMServiceAccountKey.Update().
		Where(
			bronzehistorygcpiamserviceaccountkey.ResourceID(old.ID),
			bronzehistorygcpiamserviceaccountkey.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close old history: %w", err)
	}

	// Create new history
	return h.CreateHistory(ctx, tx, new, now)
}

// CloseHistory closes history records for a deleted service account key.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPIAMServiceAccountKey.Update().
		Where(
			bronzehistorygcpiamserviceaccountkey.ResourceID(resourceID),
			bronzehistorygcpiamserviceaccountkey.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if ent.IsNotFound(err) {
		return nil // No history to close
	}
	return err
}
