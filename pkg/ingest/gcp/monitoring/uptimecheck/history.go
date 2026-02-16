package uptimecheck

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpmonitoringuptimecheckconfig"
)

// HistoryService manages uptime check config history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new uptime check config.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *UptimeCheckData, now time.Time) error {
	create := tx.BronzeHistoryGCPMonitoringUptimeCheckConfig.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetCheckerType(data.CheckerType).
		SetIsInternal(data.IsInternal).
		SetProjectID(data.ProjectID)

	if data.DisplayName != "" {
		create.SetDisplayName(data.DisplayName)
	}
	if data.Period != "" {
		create.SetPeriod(data.Period)
	}
	if data.Timeout != "" {
		create.SetTimeout(data.Timeout)
	}
	if data.MonitoredResourceJSON != nil {
		create.SetMonitoredResourceJSON(data.MonitoredResourceJSON)
	}
	if data.ResourceGroupJSON != nil {
		create.SetResourceGroupJSON(data.ResourceGroupJSON)
	}
	if data.HttpCheckJSON != nil {
		create.SetHTTPCheckJSON(data.HttpCheckJSON)
	}
	if data.TcpCheckJSON != nil {
		create.SetTCPCheckJSON(data.TcpCheckJSON)
	}
	if data.ContentMatchersJSON != nil {
		create.SetContentMatchersJSON(data.ContentMatchersJSON)
	}
	if data.SelectedRegionsJSON != nil {
		create.SetSelectedRegionsJSON(data.SelectedRegionsJSON)
	}
	if data.InternalCheckersJSON != nil {
		create.SetInternalCheckersJSON(data.InternalCheckersJSON)
	}
	if data.UserLabelsJSON != nil {
		create.SetUserLabelsJSON(data.UserLabelsJSON)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create uptime check config history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed uptime check config.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPMonitoringUptimeCheckConfig, new *UptimeCheckData, diff *UptimeCheckDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPMonitoringUptimeCheckConfig.Query().
		Where(
			bronzehistorygcpmonitoringuptimecheckconfig.ResourceID(old.ID),
			bronzehistorygcpmonitoringuptimecheckconfig.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current uptime check config history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPMonitoringUptimeCheckConfig.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current uptime check config history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPMonitoringUptimeCheckConfig.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetCheckerType(new.CheckerType).
			SetIsInternal(new.IsInternal).
			SetProjectID(new.ProjectID)

		if new.DisplayName != "" {
			create.SetDisplayName(new.DisplayName)
		}
		if new.Period != "" {
			create.SetPeriod(new.Period)
		}
		if new.Timeout != "" {
			create.SetTimeout(new.Timeout)
		}
		if new.MonitoredResourceJSON != nil {
			create.SetMonitoredResourceJSON(new.MonitoredResourceJSON)
		}
		if new.ResourceGroupJSON != nil {
			create.SetResourceGroupJSON(new.ResourceGroupJSON)
		}
		if new.HttpCheckJSON != nil {
			create.SetHTTPCheckJSON(new.HttpCheckJSON)
		}
		if new.TcpCheckJSON != nil {
			create.SetTCPCheckJSON(new.TcpCheckJSON)
		}
		if new.ContentMatchersJSON != nil {
			create.SetContentMatchersJSON(new.ContentMatchersJSON)
		}
		if new.SelectedRegionsJSON != nil {
			create.SetSelectedRegionsJSON(new.SelectedRegionsJSON)
		}
		if new.InternalCheckersJSON != nil {
			create.SetInternalCheckersJSON(new.InternalCheckersJSON)
		}
		if new.UserLabelsJSON != nil {
			create.SetUserLabelsJSON(new.UserLabelsJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new uptime check config history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted uptime check config.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPMonitoringUptimeCheckConfig.Query().
		Where(
			bronzehistorygcpmonitoringuptimecheckconfig.ResourceID(resourceID),
			bronzehistorygcpmonitoringuptimecheckconfig.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current uptime check config history: %w", err)
	}

	err = tx.BronzeHistoryGCPMonitoringUptimeCheckConfig.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close uptime check config history: %w", err)
	}

	return nil
}
