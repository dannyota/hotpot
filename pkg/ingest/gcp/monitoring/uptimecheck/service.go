package uptimecheck

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpmonitoringuptimecheckconfig"
)

// Service handles uptime check config ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new uptime check config ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for uptime check config ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of uptime check config ingestion.
type IngestResult struct {
	ProjectID        string
	UptimeCheckCount int
	CollectedAt      time.Time
	DurationMillis   int64
}

// Ingest fetches uptime check configs from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch uptime check configs from GCP
	rawConfigs, err := s.client.ListUptimeCheckConfigs(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list uptime check configs: %w", err)
	}

	// Convert to data structs
	configDataList := make([]*UptimeCheckData, 0, len(rawConfigs))
	for _, raw := range rawConfigs {
		data := ConvertUptimeCheckConfig(raw, params.ProjectID, collectedAt)
		configDataList = append(configDataList, data)
	}

	// Save to database
	if err := s.saveUptimeChecks(ctx, configDataList); err != nil {
		return nil, fmt.Errorf("failed to save uptime check configs: %w", err)
	}

	return &IngestResult{
		ProjectID:        params.ProjectID,
		UptimeCheckCount: len(configDataList),
		CollectedAt:      collectedAt,
		DurationMillis:   time.Since(startTime).Milliseconds(),
	}, nil
}

// saveUptimeChecks saves uptime check configs to the database with history tracking.
func (s *Service) saveUptimeChecks(ctx context.Context, configs []*UptimeCheckData) error {
	if len(configs) == 0 {
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

	for _, configData := range configs {
		// Load existing uptime check config
		existing, err := tx.BronzeGCPMonitoringUptimeCheckConfig.Query().
			Where(bronzegcpmonitoringuptimecheckconfig.ID(configData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing uptime check config %s: %w", configData.ID, err)
		}

		// Compute diff
		diff := DiffUptimeCheckData(existing, configData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPMonitoringUptimeCheckConfig.UpdateOneID(configData.ID).
				SetCollectedAt(configData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for uptime check config %s: %w", configData.ID, err)
			}
			continue
		}

		// Create or update uptime check config
		if existing == nil {
			create := tx.BronzeGCPMonitoringUptimeCheckConfig.Create().
				SetID(configData.ID).
				SetName(configData.Name).
				SetCheckerType(configData.CheckerType).
				SetIsInternal(configData.IsInternal).
				SetProjectID(configData.ProjectID).
				SetCollectedAt(configData.CollectedAt).
				SetFirstCollectedAt(configData.CollectedAt)

			if configData.DisplayName != "" {
				create.SetDisplayName(configData.DisplayName)
			}
			if configData.Period != "" {
				create.SetPeriod(configData.Period)
			}
			if configData.Timeout != "" {
				create.SetTimeout(configData.Timeout)
			}
			if configData.MonitoredResourceJSON != nil {
				create.SetMonitoredResourceJSON(configData.MonitoredResourceJSON)
			}
			if configData.ResourceGroupJSON != nil {
				create.SetResourceGroupJSON(configData.ResourceGroupJSON)
			}
			if configData.HttpCheckJSON != nil {
				create.SetHTTPCheckJSON(configData.HttpCheckJSON)
			}
			if configData.TcpCheckJSON != nil {
				create.SetTCPCheckJSON(configData.TcpCheckJSON)
			}
			if configData.ContentMatchersJSON != nil {
				create.SetContentMatchersJSON(configData.ContentMatchersJSON)
			}
			if configData.SelectedRegionsJSON != nil {
				create.SetSelectedRegionsJSON(configData.SelectedRegionsJSON)
			}
			if configData.InternalCheckersJSON != nil {
				create.SetInternalCheckersJSON(configData.InternalCheckersJSON)
			}
			if configData.UserLabelsJSON != nil {
				create.SetUserLabelsJSON(configData.UserLabelsJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create uptime check config %s: %w", configData.ID, err)
			}
		} else {
			update := tx.BronzeGCPMonitoringUptimeCheckConfig.UpdateOneID(configData.ID).
				SetName(configData.Name).
				SetCheckerType(configData.CheckerType).
				SetIsInternal(configData.IsInternal).
				SetProjectID(configData.ProjectID).
				SetCollectedAt(configData.CollectedAt)

			if configData.DisplayName != "" {
				update.SetDisplayName(configData.DisplayName)
			}
			if configData.Period != "" {
				update.SetPeriod(configData.Period)
			}
			if configData.Timeout != "" {
				update.SetTimeout(configData.Timeout)
			}
			if configData.MonitoredResourceJSON != nil {
				update.SetMonitoredResourceJSON(configData.MonitoredResourceJSON)
			}
			if configData.ResourceGroupJSON != nil {
				update.SetResourceGroupJSON(configData.ResourceGroupJSON)
			}
			if configData.HttpCheckJSON != nil {
				update.SetHTTPCheckJSON(configData.HttpCheckJSON)
			}
			if configData.TcpCheckJSON != nil {
				update.SetTCPCheckJSON(configData.TcpCheckJSON)
			}
			if configData.ContentMatchersJSON != nil {
				update.SetContentMatchersJSON(configData.ContentMatchersJSON)
			}
			if configData.SelectedRegionsJSON != nil {
				update.SetSelectedRegionsJSON(configData.SelectedRegionsJSON)
			}
			if configData.InternalCheckersJSON != nil {
				update.SetInternalCheckersJSON(configData.InternalCheckersJSON)
			}
			if configData.UserLabelsJSON != nil {
				update.SetUserLabelsJSON(configData.UserLabelsJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update uptime check config %s: %w", configData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, configData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for uptime check config %s: %w", configData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, configData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for uptime check config %s: %w", configData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleUptimeChecks removes uptime check configs that were not collected in the latest run.
func (s *Service) DeleteStaleUptimeChecks(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	staleConfigs, err := tx.BronzeGCPMonitoringUptimeCheckConfig.Query().
		Where(
			bronzegcpmonitoringuptimecheckconfig.ProjectID(projectID),
			bronzegcpmonitoringuptimecheckconfig.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, cfg := range staleConfigs {
		if err := s.history.CloseHistory(ctx, tx, cfg.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for uptime check config %s: %w", cfg.ID, err)
		}

		if err := tx.BronzeGCPMonitoringUptimeCheckConfig.DeleteOne(cfg).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete uptime check config %s: %w", cfg.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
