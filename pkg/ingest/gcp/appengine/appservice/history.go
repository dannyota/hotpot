package appservice

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpappengineservice"
)

// HistoryService manages App Engine service history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new App Engine service.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *ServiceData, applicationHistoryID uint, now time.Time) error {
	create := tx.BronzeHistoryGCPAppEngineService.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetApplicationHistoryID(applicationHistoryID).
		SetName(data.Name).
		SetProjectID(data.ProjectID)

	if data.SplitJSON != nil {
		create.SetSplitJSON(data.SplitJSON)
	}
	if data.LabelsJSON != nil {
		create.SetLabelsJSON(data.LabelsJSON)
	}
	if data.NetworkSettingsJSON != nil {
		create.SetNetworkSettingsJSON(data.NetworkSettingsJSON)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create App Engine service history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed App Engine service.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPAppEngineService, new *ServiceData, diff *ServiceDiff, applicationHistoryID uint, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPAppEngineService.Query().
		Where(
			bronzehistorygcpappengineservice.ResourceID(old.ID),
			bronzehistorygcpappengineservice.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current App Engine service history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPAppEngineService.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current App Engine service history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPAppEngineService.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetApplicationHistoryID(applicationHistoryID).
			SetName(new.Name).
			SetProjectID(new.ProjectID)

		if new.SplitJSON != nil {
			create.SetSplitJSON(new.SplitJSON)
		}
		if new.LabelsJSON != nil {
			create.SetLabelsJSON(new.LabelsJSON)
		}
		if new.NetworkSettingsJSON != nil {
			create.SetNetworkSettingsJSON(new.NetworkSettingsJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new App Engine service history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted App Engine service.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPAppEngineService.Query().
		Where(
			bronzehistorygcpappengineservice.ResourceID(resourceID),
			bronzehistorygcpappengineservice.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current App Engine service history: %w", err)
	}

	err = tx.BronzeHistoryGCPAppEngineService.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close App Engine service history: %w", err)
	}

	return nil
}
