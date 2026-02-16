package cryptokey

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpkmscryptokey"
)

// HistoryService handles history tracking for crypto keys.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new crypto key.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *CryptoKeyData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPKMSCryptoKey.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetPurpose(data.Purpose).
		SetCreateTime(data.CreateTime).
		SetNextRotationTime(data.NextRotationTime).
		SetRotationPeriod(data.RotationPeriod).
		SetDestroyScheduledDuration(data.DestroyScheduledDuration).
		SetImportOnly(data.ImportOnly).
		SetCryptoKeyBackend(data.CryptoKeyBackend).
		SetVersionTemplateJSON(data.VersionTemplateJSON).
		SetPrimaryJSON(data.PrimaryJSON).
		SetLabelsJSON(data.LabelsJSON).
		SetProjectID(data.ProjectID).
		SetLocation(data.Location).
		SetKeyRingName(data.KeyRingName).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create crypto key history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPKMSCryptoKey, new *CryptoKeyData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGCPKMSCryptoKey.Query().
		Where(
			bronzehistorygcpkmscryptokey.ResourceID(old.ID),
			bronzehistorygcpkmscryptokey.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current crypto key history: %w", err)
	}

	err = tx.BronzeHistoryGCPKMSCryptoKey.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close current crypto key history: %w", err)
	}

	_, err = tx.BronzeHistoryGCPKMSCryptoKey.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetPurpose(new.Purpose).
		SetCreateTime(new.CreateTime).
		SetNextRotationTime(new.NextRotationTime).
		SetRotationPeriod(new.RotationPeriod).
		SetDestroyScheduledDuration(new.DestroyScheduledDuration).
		SetImportOnly(new.ImportOnly).
		SetCryptoKeyBackend(new.CryptoKeyBackend).
		SetVersionTemplateJSON(new.VersionTemplateJSON).
		SetPrimaryJSON(new.PrimaryJSON).
		SetLabelsJSON(new.LabelsJSON).
		SetProjectID(new.ProjectID).
		SetLocation(new.Location).
		SetKeyRingName(new.KeyRingName).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create new crypto key history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted crypto key.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGCPKMSCryptoKey.Query().
		Where(
			bronzehistorygcpkmscryptokey.ResourceID(resourceID),
			bronzehistorygcpkmscryptokey.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current crypto key history: %w", err)
	}

	err = tx.BronzeHistoryGCPKMSCryptoKey.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close crypto key history: %w", err)
	}

	return nil
}
