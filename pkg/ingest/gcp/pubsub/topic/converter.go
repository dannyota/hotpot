package topic

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub/apiv1/pubsubpb"
)

// TopicData holds converted topic data ready for Ent insertion.
type TopicData struct {
	ID                              string
	Name                            string
	LabelsJSON                      json.RawMessage
	MessageStoragePolicyJSON        json.RawMessage
	KmsKeyName                      string
	SchemaSettingsJSON              json.RawMessage
	MessageRetentionDuration        string
	State                           int
	IngestionDataSourceSettingsJSON json.RawMessage
	ProjectID                       string
	CollectedAt                     time.Time
}

// ConvertTopic converts a raw GCP API Pub/Sub topic to Ent-compatible data.
func ConvertTopic(t *pubsubpb.Topic, projectID string, collectedAt time.Time) (*TopicData, error) {
	data := &TopicData{
		ID:                       t.GetName(),
		Name:                     t.GetName(),
		KmsKeyName:               t.GetKmsKeyName(),
		MessageRetentionDuration: t.GetMessageRetentionDuration().String(),
		State:                    int(t.GetState()),
		ProjectID:                projectID,
		CollectedAt:              collectedAt,
	}

	// Handle "0s" as empty (proto default)
	if t.GetMessageRetentionDuration().String() == "0s" || t.MessageRetentionDuration == nil {
		data.MessageRetentionDuration = ""
	}

	// Convert labels to JSON
	if len(t.GetLabels()) > 0 {
		labelsJSON, err := json.Marshal(t.GetLabels())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal labels for topic %s: %w", t.GetName(), err)
		}
		data.LabelsJSON = labelsJSON
	}

	// Convert message storage policy to JSON
	if t.MessageStoragePolicy != nil {
		policyJSON, err := json.Marshal(t.MessageStoragePolicy)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal message storage policy for topic %s: %w", t.GetName(), err)
		}
		data.MessageStoragePolicyJSON = policyJSON
	}

	// Convert schema settings to JSON
	if t.SchemaSettings != nil {
		settingsJSON, err := json.Marshal(t.SchemaSettings)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal schema settings for topic %s: %w", t.GetName(), err)
		}
		data.SchemaSettingsJSON = settingsJSON
	}

	// Convert ingestion data source settings to JSON
	if t.IngestionDataSourceSettings != nil {
		settingsJSON, err := json.Marshal(t.IngestionDataSourceSettings)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal ingestion data source settings for topic %s: %w", t.GetName(), err)
		}
		data.IngestionDataSourceSettingsJSON = settingsJSON
	}

	return data, nil
}
