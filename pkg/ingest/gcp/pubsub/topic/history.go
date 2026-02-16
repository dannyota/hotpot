package topic

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcppubsubtopic"
)

// HistoryService manages Pub/Sub topic history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new topic.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *TopicData, now time.Time) error {
	create := tx.BronzeHistoryGCPPubSubTopic.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetKmsKeyName(data.KmsKeyName).
		SetMessageRetentionDuration(data.MessageRetentionDuration).
		SetState(data.State).
		SetProjectID(data.ProjectID)

	if data.LabelsJSON != nil {
		create.SetLabelsJSON(data.LabelsJSON)
	}
	if data.MessageStoragePolicyJSON != nil {
		create.SetMessageStoragePolicyJSON(data.MessageStoragePolicyJSON)
	}
	if data.SchemaSettingsJSON != nil {
		create.SetSchemaSettingsJSON(data.SchemaSettingsJSON)
	}
	if data.IngestionDataSourceSettingsJSON != nil {
		create.SetIngestionDataSourceSettingsJSON(data.IngestionDataSourceSettingsJSON)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create topic history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed topic.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPPubSubTopic, new *TopicData, diff *TopicDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPPubSubTopic.Query().
		Where(
			bronzehistorygcppubsubtopic.ResourceID(old.ID),
			bronzehistorygcppubsubtopic.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current topic history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPPubSubTopic.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current topic history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPPubSubTopic.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetKmsKeyName(new.KmsKeyName).
			SetMessageRetentionDuration(new.MessageRetentionDuration).
			SetState(new.State).
			SetProjectID(new.ProjectID)

		if new.LabelsJSON != nil {
			create.SetLabelsJSON(new.LabelsJSON)
		}
		if new.MessageStoragePolicyJSON != nil {
			create.SetMessageStoragePolicyJSON(new.MessageStoragePolicyJSON)
		}
		if new.SchemaSettingsJSON != nil {
			create.SetSchemaSettingsJSON(new.SchemaSettingsJSON)
		}
		if new.IngestionDataSourceSettingsJSON != nil {
			create.SetIngestionDataSourceSettingsJSON(new.IngestionDataSourceSettingsJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new topic history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted topic.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPPubSubTopic.Query().
		Where(
			bronzehistorygcppubsubtopic.ResourceID(resourceID),
			bronzehistorygcppubsubtopic.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current topic history: %w", err)
	}

	err = tx.BronzeHistoryGCPPubSubTopic.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close topic history: %w", err)
	}

	return nil
}
