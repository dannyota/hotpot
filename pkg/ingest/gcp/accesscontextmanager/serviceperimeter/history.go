package serviceperimeter

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpaccesscontextmanagerserviceperimeter"
)

// HistoryService manages service perimeter history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new service perimeter.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *ServicePerimeterData, now time.Time) error {
	create := tx.BronzeHistoryGCPAccessContextManagerServicePerimeter.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetPerimeterType(data.PerimeterType).
		SetUseExplicitDryRunSpec(data.UseExplicitDryRunSpec).
		SetAccessPolicyName(data.AccessPolicyName).
		SetOrganizationID(data.OrganizationID)

	if data.Title != "" {
		create.SetTitle(data.Title)
	}
	if data.Description != "" {
		create.SetDescription(data.Description)
	}
	if data.Etag != "" {
		create.SetEtag(data.Etag)
	}
	if data.StatusJSON != nil {
		create.SetStatusJSON(data.StatusJSON)
	}
	if data.SpecJSON != nil {
		create.SetSpecJSON(data.SpecJSON)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create service perimeter history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed service perimeter.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPAccessContextManagerServicePerimeter, new *ServicePerimeterData, diff *ServicePerimeterDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPAccessContextManagerServicePerimeter.Query().
		Where(
			bronzehistorygcpaccesscontextmanagerserviceperimeter.ResourceID(old.ID),
			bronzehistorygcpaccesscontextmanagerserviceperimeter.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current service perimeter history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPAccessContextManagerServicePerimeter.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current service perimeter history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPAccessContextManagerServicePerimeter.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetPerimeterType(new.PerimeterType).
			SetUseExplicitDryRunSpec(new.UseExplicitDryRunSpec).
			SetAccessPolicyName(new.AccessPolicyName).
			SetOrganizationID(new.OrganizationID)

		if new.Title != "" {
			create.SetTitle(new.Title)
		}
		if new.Description != "" {
			create.SetDescription(new.Description)
		}
		if new.Etag != "" {
			create.SetEtag(new.Etag)
		}
		if new.StatusJSON != nil {
			create.SetStatusJSON(new.StatusJSON)
		}
		if new.SpecJSON != nil {
			create.SetSpecJSON(new.SpecJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new service perimeter history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted service perimeter.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPAccessContextManagerServicePerimeter.Query().
		Where(
			bronzehistorygcpaccesscontextmanagerserviceperimeter.ResourceID(resourceID),
			bronzehistorygcpaccesscontextmanagerserviceperimeter.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current service perimeter history: %w", err)
	}

	err = tx.BronzeHistoryGCPAccessContextManagerServicePerimeter.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close service perimeter history: %w", err)
	}

	return nil
}
