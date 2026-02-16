package enabledservice

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpserviceusageenabledservice"
)

// HistoryService manages enabled service history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new enabled service.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *EnabledServiceData, now time.Time) error {
	create := tx.BronzeHistoryGCPServiceUsageEnabledService.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetParent(data.Parent).
		SetState(data.State).
		SetProjectID(data.ProjectID)

	if data.ConfigJSON != nil {
		create.SetConfigJSON(data.ConfigJSON)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create enabled service history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed enabled service.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPServiceUsageEnabledService, new *EnabledServiceData, diff *EnabledServiceDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPServiceUsageEnabledService.Query().
		Where(
			bronzehistorygcpserviceusageenabledservice.ResourceID(old.ID),
			bronzehistorygcpserviceusageenabledservice.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current enabled service history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPServiceUsageEnabledService.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current enabled service history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPServiceUsageEnabledService.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetParent(new.Parent).
			SetState(new.State).
			SetProjectID(new.ProjectID)

		if new.ConfigJSON != nil {
			create.SetConfigJSON(new.ConfigJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new enabled service history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted enabled service.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPServiceUsageEnabledService.Query().
		Where(
			bronzehistorygcpserviceusageenabledservice.ResourceID(resourceID),
			bronzehistorygcpserviceusageenabledservice.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current enabled service history: %w", err)
	}

	err = tx.BronzeHistoryGCPServiceUsageEnabledService.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close enabled service history: %w", err)
	}

	return nil
}
