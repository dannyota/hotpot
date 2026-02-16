package key

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorydokey"
)

// HistoryService handles history tracking for SSH keys.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) buildCreate(tx *ent.Tx, data *KeyData) *ent.BronzeHistoryDOKeyCreate {
	return tx.BronzeHistoryDOKey.Create().
		SetResourceID(data.ResourceID).
		SetName(data.Name).
		SetFingerprint(data.Fingerprint).
		SetPublicKey(data.PublicKey)
}

// CreateHistory creates a history record for a new SSH key.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *KeyData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create key history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new for a changed SSH key.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeDOKey, new *KeyData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDOKey.Query().
		Where(
			bronzehistorydokey.ResourceID(old.ID),
			bronzehistorydokey.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current key history: %w", err)
	}

	if err := tx.BronzeHistoryDOKey.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close key history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new key history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted SSH key.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDOKey.Query().
		Where(
			bronzehistorydokey.ResourceID(resourceID),
			bronzehistorydokey.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current key history: %w", err)
	}

	if err := tx.BronzeHistoryDOKey.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close key history: %w", err)
	}

	return nil
}
