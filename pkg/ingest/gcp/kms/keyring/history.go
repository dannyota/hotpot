package keyring

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpkmskeyring"
)

// HistoryService handles history tracking for key rings.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new key ring.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *KeyRingData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPKMSKeyRing.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetCreateTime(data.CreateTime).
		SetProjectID(data.ProjectID).
		SetLocation(data.Location).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create key ring history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPKMSKeyRing, new *KeyRingData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGCPKMSKeyRing.Query().
		Where(
			bronzehistorygcpkmskeyring.ResourceID(old.ID),
			bronzehistorygcpkmskeyring.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current key ring history: %w", err)
	}

	err = tx.BronzeHistoryGCPKMSKeyRing.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close current key ring history: %w", err)
	}

	_, err = tx.BronzeHistoryGCPKMSKeyRing.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetCreateTime(new.CreateTime).
		SetProjectID(new.ProjectID).
		SetLocation(new.Location).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create new key ring history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted key ring.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGCPKMSKeyRing.Query().
		Where(
			bronzehistorygcpkmskeyring.ResourceID(resourceID),
			bronzehistorygcpkmskeyring.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current key ring history: %w", err)
	}

	err = tx.BronzeHistoryGCPKMSKeyRing.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close key ring history: %w", err)
	}

	return nil
}
