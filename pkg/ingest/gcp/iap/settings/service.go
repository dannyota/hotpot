package settings

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpiapsettings"
)

// Service handles IAP settings ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new IAP settings ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for IAP settings ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of IAP settings ingestion.
type IngestResult struct {
	ProjectID      string
	SettingsCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches IAP settings from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch IAP settings from GCP (single object per project)
	raw, err := s.client.GetSettings(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get IAP settings: %w", err)
	}

	// Convert to data struct
	data, err := ConvertSettings(raw, params.ProjectID, collectedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to convert IAP settings: %w", err)
	}

	// Save to database
	dataList := []*SettingsData{data}
	if err := s.saveSettings(ctx, dataList); err != nil {
		return nil, fmt.Errorf("failed to save IAP settings: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		SettingsCount:  1,
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveSettings saves IAP settings to the database with history tracking.
func (s *Service) saveSettings(ctx context.Context, settingsList []*SettingsData) error {
	if len(settingsList) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, data := range settingsList {
		// Load existing settings
		existing, err := tx.BronzeGCPIAPSettings.Query().
			Where(bronzegcpiapsettings.ID(data.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing IAP settings %s: %w", data.ID, err)
		}

		// Compute diff
		diff := DiffSettingsData(existing, data)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPIAPSettings.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for IAP settings %s: %w", data.ID, err)
			}
			continue
		}

		// Create or update settings
		if existing == nil {
			create := tx.BronzeGCPIAPSettings.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.AccessSettingsJSON != nil {
				create.SetAccessSettingsJSON(data.AccessSettingsJSON)
			}
			if data.ApplicationSettingsJSON != nil {
				create.SetApplicationSettingsJSON(data.ApplicationSettingsJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create IAP settings %s: %w", data.ID, err)
			}
		} else {
			update := tx.BronzeGCPIAPSettings.UpdateOneID(data.ID).
				SetName(data.Name).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt)

			if data.AccessSettingsJSON != nil {
				update.SetAccessSettingsJSON(data.AccessSettingsJSON)
			}
			if data.ApplicationSettingsJSON != nil {
				update.SetApplicationSettingsJSON(data.ApplicationSettingsJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update IAP settings %s: %w", data.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for IAP settings %s: %w", data.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, data, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for IAP settings %s: %w", data.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleSettings removes IAP settings that were not collected in the latest run.
func (s *Service) DeleteStaleSettings(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	staleSettings, err := tx.BronzeGCPIAPSettings.Query().
		Where(
			bronzegcpiapsettings.ProjectID(projectID),
			bronzegcpiapsettings.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, stale := range staleSettings {
		if err := s.history.CloseHistory(ctx, tx, stale.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for IAP settings %s: %w", stale.ID, err)
		}

		if err := tx.BronzeGCPIAPSettings.DeleteOne(stale).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete IAP settings %s: %w", stale.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
