package notificationconfig

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpsecuritycenternotificationconfig"
)

// Service handles SCC notification config ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new SCC notification config ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of SCC notification config ingestion.
type IngestResult struct {
	NotificationConfigCount int
	CollectedAt             time.Time
	DurationMillis          int64
}

// Ingest fetches SCC notification configs from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch notification configs from GCP
	rawConfigs, err := s.client.ListNotificationConfigs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list notification configs: %w", err)
	}

	// Convert to notification config data
	configDataList := make([]*NotificationConfigData, 0, len(rawConfigs))
	for _, raw := range rawConfigs {
		data := ConvertNotificationConfig(raw.OrgName, raw.NotificationConfig, collectedAt)
		configDataList = append(configDataList, data)
	}

	// Save to database
	if err := s.saveNotificationConfigs(ctx, configDataList); err != nil {
		return nil, fmt.Errorf("failed to save notification configs: %w", err)
	}

	return &IngestResult{
		NotificationConfigCount: len(configDataList),
		CollectedAt:             collectedAt,
		DurationMillis:          time.Since(startTime).Milliseconds(),
	}, nil
}

// saveNotificationConfigs saves SCC notification configs to the database with history tracking.
func (s *Service) saveNotificationConfigs(ctx context.Context, configs []*NotificationConfigData) error {
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
		// Load existing notification config
		existing, err := tx.BronzeGCPSecurityCenterNotificationConfig.Query().
			Where(bronzegcpsecuritycenternotificationconfig.ID(configData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing notification config %s: %w", configData.ID, err)
		}

		// Compute diff
		diff := DiffNotificationConfigData(existing, configData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPSecurityCenterNotificationConfig.UpdateOneID(configData.ID).
				SetCollectedAt(configData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for notification config %s: %w", configData.ID, err)
			}
			continue
		}

		// Create or update notification config
		if existing == nil {
			create := tx.BronzeGCPSecurityCenterNotificationConfig.Create().
				SetID(configData.ID).
				SetName(configData.Name).
				SetOrganizationID(configData.OrganizationID).
				SetCollectedAt(configData.CollectedAt).
				SetFirstCollectedAt(configData.CollectedAt)

			if configData.Description != "" {
				create.SetDescription(configData.Description)
			}
			if configData.PubsubTopic != "" {
				create.SetPubsubTopic(configData.PubsubTopic)
			}
			if configData.StreamingConfigJSON != "" {
				create.SetStreamingConfigJSON(configData.StreamingConfigJSON)
			}
			if configData.ServiceAccount != "" {
				create.SetServiceAccount(configData.ServiceAccount)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create notification config %s: %w", configData.ID, err)
			}
		} else {
			update := tx.BronzeGCPSecurityCenterNotificationConfig.UpdateOneID(configData.ID).
				SetName(configData.Name).
				SetOrganizationID(configData.OrganizationID).
				SetCollectedAt(configData.CollectedAt)

			if configData.Description != "" {
				update.SetDescription(configData.Description)
			}
			if configData.PubsubTopic != "" {
				update.SetPubsubTopic(configData.PubsubTopic)
			}
			if configData.StreamingConfigJSON != "" {
				update.SetStreamingConfigJSON(configData.StreamingConfigJSON)
			}
			if configData.ServiceAccount != "" {
				update.SetServiceAccount(configData.ServiceAccount)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update notification config %s: %w", configData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, configData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for notification config %s: %w", configData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, configData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for notification config %s: %w", configData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleNotificationConfigs removes notification configs that were not collected in the latest run.
func (s *Service) DeleteStaleNotificationConfigs(ctx context.Context, collectedAt time.Time) error {
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

	staleConfigs, err := tx.BronzeGCPSecurityCenterNotificationConfig.Query().
		Where(bronzegcpsecuritycenternotificationconfig.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, config := range staleConfigs {
		if err := s.history.CloseHistory(ctx, tx, config.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for notification config %s: %w", config.ID, err)
		}

		if err := tx.BronzeGCPSecurityCenterNotificationConfig.DeleteOne(config).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete notification config %s: %w", config.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
