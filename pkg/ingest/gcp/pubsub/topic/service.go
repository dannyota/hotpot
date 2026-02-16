package topic

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcppubsubtopic"
)

// Service handles Pub/Sub topic ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new topic ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for topic ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of topic ingestion.
type IngestResult struct {
	ProjectID      string
	TopicCount     int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches topics from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch topics from GCP
	rawTopics, err := s.client.ListTopics(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list topics: %w", err)
	}

	// Convert to topic data
	topicDataList := make([]*TopicData, 0, len(rawTopics))
	for _, raw := range rawTopics {
		data, err := ConvertTopic(raw, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert topic: %w", err)
		}
		topicDataList = append(topicDataList, data)
	}

	// Save to database
	if err := s.saveTopics(ctx, topicDataList); err != nil {
		return nil, fmt.Errorf("failed to save topics: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		TopicCount:     len(topicDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveTopics saves topics to the database with history tracking.
func (s *Service) saveTopics(ctx context.Context, topics []*TopicData) error {
	if len(topics) == 0 {
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

	for _, topicData := range topics {
		// Load existing topic
		existing, err := tx.BronzeGCPPubSubTopic.Query().
			Where(bronzegcppubsubtopic.ID(topicData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing topic %s: %w", topicData.ID, err)
		}

		// Compute diff
		diff := DiffTopicData(existing, topicData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPPubSubTopic.UpdateOneID(topicData.ID).
				SetCollectedAt(topicData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for topic %s: %w", topicData.ID, err)
			}
			continue
		}

		// Create or update topic
		if existing == nil {
			create := tx.BronzeGCPPubSubTopic.Create().
				SetID(topicData.ID).
				SetName(topicData.Name).
				SetProjectID(topicData.ProjectID).
				SetCollectedAt(topicData.CollectedAt).
				SetFirstCollectedAt(topicData.CollectedAt).
				SetState(topicData.State)

			if topicData.KmsKeyName != "" {
				create.SetKmsKeyName(topicData.KmsKeyName)
			}
			if topicData.MessageRetentionDuration != "" {
				create.SetMessageRetentionDuration(topicData.MessageRetentionDuration)
			}
			if topicData.LabelsJSON != nil {
				create.SetLabelsJSON(topicData.LabelsJSON)
			}
			if topicData.MessageStoragePolicyJSON != nil {
				create.SetMessageStoragePolicyJSON(topicData.MessageStoragePolicyJSON)
			}
			if topicData.SchemaSettingsJSON != nil {
				create.SetSchemaSettingsJSON(topicData.SchemaSettingsJSON)
			}
			if topicData.IngestionDataSourceSettingsJSON != nil {
				create.SetIngestionDataSourceSettingsJSON(topicData.IngestionDataSourceSettingsJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create topic %s: %w", topicData.ID, err)
			}
		} else {
			update := tx.BronzeGCPPubSubTopic.UpdateOneID(topicData.ID).
				SetName(topicData.Name).
				SetProjectID(topicData.ProjectID).
				SetCollectedAt(topicData.CollectedAt).
				SetState(topicData.State)

			if topicData.KmsKeyName != "" {
				update.SetKmsKeyName(topicData.KmsKeyName)
			}
			if topicData.MessageRetentionDuration != "" {
				update.SetMessageRetentionDuration(topicData.MessageRetentionDuration)
			}
			if topicData.LabelsJSON != nil {
				update.SetLabelsJSON(topicData.LabelsJSON)
			}
			if topicData.MessageStoragePolicyJSON != nil {
				update.SetMessageStoragePolicyJSON(topicData.MessageStoragePolicyJSON)
			}
			if topicData.SchemaSettingsJSON != nil {
				update.SetSchemaSettingsJSON(topicData.SchemaSettingsJSON)
			}
			if topicData.IngestionDataSourceSettingsJSON != nil {
				update.SetIngestionDataSourceSettingsJSON(topicData.IngestionDataSourceSettingsJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update topic %s: %w", topicData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, topicData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for topic %s: %w", topicData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, topicData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for topic %s: %w", topicData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleTopics removes topics that were not collected in the latest run.
func (s *Service) DeleteStaleTopics(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	staleTopics, err := tx.BronzeGCPPubSubTopic.Query().
		Where(
			bronzegcppubsubtopic.ProjectID(projectID),
			bronzegcppubsubtopic.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, t := range staleTopics {
		if err := s.history.CloseHistory(ctx, tx, t.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for topic %s: %w", t.ID, err)
		}

		if err := tx.BronzeGCPPubSubTopic.DeleteOne(t).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete topic %s: %w", t.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
