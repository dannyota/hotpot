package settings

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpiapsettings"
)

// HistoryService manages IAP settings history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for new IAP settings.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *SettingsData, now time.Time) error {
	create := tx.BronzeHistoryGCPIAPSettings.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetProjectID(data.ProjectID)

	if data.AccessSettingsJSON != nil {
		create.SetAccessSettingsJSON(data.AccessSettingsJSON)
	}
	if data.ApplicationSettingsJSON != nil {
		create.SetApplicationSettingsJSON(data.ApplicationSettingsJSON)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create IAP settings history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for changed IAP settings.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPIAPSettings, new *SettingsData, diff *SettingsDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPIAPSettings.Query().
		Where(
			bronzehistorygcpiapsettings.ResourceID(old.ID),
			bronzehistorygcpiapsettings.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current IAP settings history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPIAPSettings.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current IAP settings history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPIAPSettings.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetProjectID(new.ProjectID)

		if new.AccessSettingsJSON != nil {
			create.SetAccessSettingsJSON(new.AccessSettingsJSON)
		}
		if new.ApplicationSettingsJSON != nil {
			create.SetApplicationSettingsJSON(new.ApplicationSettingsJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new IAP settings history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for deleted IAP settings.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPIAPSettings.Query().
		Where(
			bronzehistorygcpiapsettings.ResourceID(resourceID),
			bronzehistorygcpiapsettings.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current IAP settings history: %w", err)
	}

	err = tx.BronzeHistoryGCPIAPSettings.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close IAP settings history: %w", err)
	}

	return nil
}
