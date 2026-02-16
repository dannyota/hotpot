package alertpolicy

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpmonitoringalertpolicy"
)

// HistoryService manages alert policy history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new alert policy.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *AlertPolicyData, now time.Time) error {
	create := tx.BronzeHistoryGCPMonitoringAlertPolicy.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetCombiner(data.Combiner).
		SetEnabled(data.Enabled).
		SetSeverity(data.Severity).
		SetProjectID(data.ProjectID)

	if data.DisplayName != "" {
		create.SetDisplayName(data.DisplayName)
	}
	if data.DocumentationJSON != nil {
		create.SetDocumentationJSON(data.DocumentationJSON)
	}
	if data.UserLabelsJSON != nil {
		create.SetUserLabelsJSON(data.UserLabelsJSON)
	}
	if data.ConditionsJSON != nil {
		create.SetConditionsJSON(data.ConditionsJSON)
	}
	if data.NotificationChannelsJSON != nil {
		create.SetNotificationChannelsJSON(data.NotificationChannelsJSON)
	}
	if data.CreationRecordJSON != nil {
		create.SetCreationRecordJSON(data.CreationRecordJSON)
	}
	if data.MutationRecordJSON != nil {
		create.SetMutationRecordJSON(data.MutationRecordJSON)
	}
	if data.AlertStrategyJSON != nil {
		create.SetAlertStrategyJSON(data.AlertStrategyJSON)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create alert policy history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed alert policy.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPMonitoringAlertPolicy, new *AlertPolicyData, diff *AlertPolicyDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPMonitoringAlertPolicy.Query().
		Where(
			bronzehistorygcpmonitoringalertpolicy.ResourceID(old.ID),
			bronzehistorygcpmonitoringalertpolicy.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current alert policy history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPMonitoringAlertPolicy.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current alert policy history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPMonitoringAlertPolicy.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetCombiner(new.Combiner).
			SetEnabled(new.Enabled).
			SetSeverity(new.Severity).
			SetProjectID(new.ProjectID)

		if new.DisplayName != "" {
			create.SetDisplayName(new.DisplayName)
		}
		if new.DocumentationJSON != nil {
			create.SetDocumentationJSON(new.DocumentationJSON)
		}
		if new.UserLabelsJSON != nil {
			create.SetUserLabelsJSON(new.UserLabelsJSON)
		}
		if new.ConditionsJSON != nil {
			create.SetConditionsJSON(new.ConditionsJSON)
		}
		if new.NotificationChannelsJSON != nil {
			create.SetNotificationChannelsJSON(new.NotificationChannelsJSON)
		}
		if new.CreationRecordJSON != nil {
			create.SetCreationRecordJSON(new.CreationRecordJSON)
		}
		if new.MutationRecordJSON != nil {
			create.SetMutationRecordJSON(new.MutationRecordJSON)
		}
		if new.AlertStrategyJSON != nil {
			create.SetAlertStrategyJSON(new.AlertStrategyJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new alert policy history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted alert policy.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPMonitoringAlertPolicy.Query().
		Where(
			bronzehistorygcpmonitoringalertpolicy.ResourceID(resourceID),
			bronzehistorygcpmonitoringalertpolicy.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current alert policy history: %w", err)
	}

	err = tx.BronzeHistoryGCPMonitoringAlertPolicy.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close alert policy history: %w", err)
	}

	return nil
}
