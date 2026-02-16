package attestor

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpbinaryauthorizationattestor"
)

// HistoryService manages Binary Authorization attestor history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new Binary Authorization attestor.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *AttestorData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPBinaryAuthorizationAttestor.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetDescription(data.Description).
		SetUserOwnedGrafeasNoteJSON(data.UserOwnedGrafeasNoteJSON).
		SetUpdateTime(data.UpdateTime).
		SetEtag(data.Etag).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create binary authorization attestor history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed Binary Authorization attestor.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPBinaryAuthorizationAttestor, new *AttestorData, diff *AttestorDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPBinaryAuthorizationAttestor.Query().
		Where(
			bronzehistorygcpbinaryauthorizationattestor.ResourceID(old.ID),
			bronzehistorygcpbinaryauthorizationattestor.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current binary authorization attestor history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPBinaryAuthorizationAttestor.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current binary authorization attestor history: %w", err)
		}

		// Create new history
		_, err := tx.BronzeHistoryGCPBinaryAuthorizationAttestor.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetDescription(new.Description).
			SetUserOwnedGrafeasNoteJSON(new.UserOwnedGrafeasNoteJSON).
			SetUpdateTime(new.UpdateTime).
			SetEtag(new.Etag).
			SetProjectID(new.ProjectID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new binary authorization attestor history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted Binary Authorization attestor.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPBinaryAuthorizationAttestor.Query().
		Where(
			bronzehistorygcpbinaryauthorizationattestor.ResourceID(resourceID),
			bronzehistorygcpbinaryauthorizationattestor.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current binary authorization attestor history: %w", err)
	}

	err = tx.BronzeHistoryGCPBinaryAuthorizationAttestor.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close binary authorization attestor history: %w", err)
	}

	return nil
}
