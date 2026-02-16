package accesspolicy

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpaccesscontextmanageraccesspolicy"
)

// HistoryService manages access policy history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new access policy.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *AccessPolicyData, now time.Time) error {
	create := tx.BronzeHistoryGCPAccessContextManagerAccessPolicy.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetParent(data.Parent).
		SetOrganizationID(data.OrganizationID)

	if data.Title != "" {
		create.SetTitle(data.Title)
	}
	if data.Etag != "" {
		create.SetEtag(data.Etag)
	}
	if data.ScopesJSON != nil {
		create.SetScopesJSON(data.ScopesJSON)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create access policy history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed access policy.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPAccessContextManagerAccessPolicy, new *AccessPolicyData, diff *AccessPolicyDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPAccessContextManagerAccessPolicy.Query().
		Where(
			bronzehistorygcpaccesscontextmanageraccesspolicy.ResourceID(old.ID),
			bronzehistorygcpaccesscontextmanageraccesspolicy.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current access policy history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPAccessContextManagerAccessPolicy.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current access policy history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPAccessContextManagerAccessPolicy.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetParent(new.Parent).
			SetOrganizationID(new.OrganizationID)

		if new.Title != "" {
			create.SetTitle(new.Title)
		}
		if new.Etag != "" {
			create.SetEtag(new.Etag)
		}
		if new.ScopesJSON != nil {
			create.SetScopesJSON(new.ScopesJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new access policy history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted access policy.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPAccessContextManagerAccessPolicy.Query().
		Where(
			bronzehistorygcpaccesscontextmanageraccesspolicy.ResourceID(resourceID),
			bronzehistorygcpaccesscontextmanageraccesspolicy.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current access policy history: %w", err)
	}

	err = tx.BronzeHistoryGCPAccessContextManagerAccessPolicy.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close access policy history: %w", err)
	}

	return nil
}
