package accesslevel

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpaccesscontextmanageraccesslevel"
)

// HistoryService manages access level history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new access level.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *AccessLevelData, now time.Time) error {
	create := tx.BronzeHistoryGCPAccessContextManagerAccessLevel.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetAccessPolicyName(data.AccessPolicyName).
		SetOrganizationID(data.OrganizationID)

	if data.Title != "" {
		create.SetTitle(data.Title)
	}
	if data.Description != "" {
		create.SetDescription(data.Description)
	}
	if data.BasicJSON != nil {
		create.SetBasicJSON(data.BasicJSON)
	}
	if data.CustomJSON != nil {
		create.SetCustomJSON(data.CustomJSON)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create access level history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed access level.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPAccessContextManagerAccessLevel, new *AccessLevelData, diff *AccessLevelDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPAccessContextManagerAccessLevel.Query().
		Where(
			bronzehistorygcpaccesscontextmanageraccesslevel.ResourceID(old.ID),
			bronzehistorygcpaccesscontextmanageraccesslevel.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current access level history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPAccessContextManagerAccessLevel.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current access level history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPAccessContextManagerAccessLevel.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetAccessPolicyName(new.AccessPolicyName).
			SetOrganizationID(new.OrganizationID)

		if new.Title != "" {
			create.SetTitle(new.Title)
		}
		if new.Description != "" {
			create.SetDescription(new.Description)
		}
		if new.BasicJSON != nil {
			create.SetBasicJSON(new.BasicJSON)
		}
		if new.CustomJSON != nil {
			create.SetCustomJSON(new.CustomJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new access level history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted access level.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPAccessContextManagerAccessLevel.Query().
		Where(
			bronzehistorygcpaccesscontextmanageraccesslevel.ResourceID(resourceID),
			bronzehistorygcpaccesscontextmanageraccesslevel.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current access level history: %w", err)
	}

	err = tx.BronzeHistoryGCPAccessContextManagerAccessLevel.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close access level history: %w", err)
	}

	return nil
}
