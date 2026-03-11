package subscription

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub/apiv1/pubsubpb"
)

// SubscriptionData holds converted subscription data ready for Ent insertion.
type SubscriptionData struct {
	ID                       string
	Name                     string
	Topic                    string
	PushConfigJSON           json.RawMessage
	BigqueryConfigJSON       json.RawMessage
	CloudStorageConfigJSON   json.RawMessage
	AckDeadlineSeconds       int
	RetainAckedMessages      bool
	MessageRetentionDuration string
	LabelsJSON               json.RawMessage
	EnableMessageOrdering    bool
	ExpirationPolicyJSON     json.RawMessage
	Filter                   string
	DeadLetterPolicyJSON     json.RawMessage
	RetryPolicyJSON          json.RawMessage
	Detached                 bool
	EnableExactlyOnceDelivery bool
	State                    int
	ProjectID                string
	CollectedAt              time.Time
}

// ConvertSubscription converts a raw GCP API Pub/Sub subscription to Ent-compatible data.
func ConvertSubscription(s *pubsubpb.Subscription, projectID string, collectedAt time.Time) (*SubscriptionData, error) {
	data := &SubscriptionData{
		ID:                        s.GetName(),
		Name:                      s.GetName(),
		Topic:                     s.GetTopic(),
		AckDeadlineSeconds:        int(s.GetAckDeadlineSeconds()),
		RetainAckedMessages:       s.GetRetainAckedMessages(),
		EnableMessageOrdering:     s.GetEnableMessageOrdering(),
		Filter:                    s.GetFilter(),
		Detached:                  s.GetDetached(),
		EnableExactlyOnceDelivery: s.GetEnableExactlyOnceDelivery(),
		State:                     int(s.GetState()),
		ProjectID:                 projectID,
		CollectedAt:               collectedAt,
	}

	// Handle message retention duration
	if s.MessageRetentionDuration != nil && s.GetMessageRetentionDuration().String() != "0s" {
		data.MessageRetentionDuration = s.GetMessageRetentionDuration().String()
	}

	// Convert labels to JSON
	if len(s.GetLabels()) > 0 {
		labelsJSON, err := json.Marshal(s.GetLabels())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal labels for subscription %s: %w", s.GetName(), err)
		}
		data.LabelsJSON = labelsJSON
	}

	// Convert push config to JSON
	if s.PushConfig != nil {
		configJSON, err := json.Marshal(s.PushConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal push config for subscription %s: %w", s.GetName(), err)
		}
		data.PushConfigJSON = configJSON
	}

	// Convert BigQuery config to JSON
	if s.BigqueryConfig != nil {
		configJSON, err := json.Marshal(s.BigqueryConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal bigquery config for subscription %s: %w", s.GetName(), err)
		}
		data.BigqueryConfigJSON = configJSON
	}

	// Convert Cloud Storage config to JSON
	if s.CloudStorageConfig != nil {
		configJSON, err := json.Marshal(s.CloudStorageConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal cloud storage config for subscription %s: %w", s.GetName(), err)
		}
		data.CloudStorageConfigJSON = configJSON
	}

	// Convert expiration policy to JSON
	if s.ExpirationPolicy != nil {
		policyJSON, err := json.Marshal(s.ExpirationPolicy)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal expiration policy for subscription %s: %w", s.GetName(), err)
		}
		data.ExpirationPolicyJSON = policyJSON
	}

	// Convert dead letter policy to JSON
	if s.DeadLetterPolicy != nil {
		policyJSON, err := json.Marshal(s.DeadLetterPolicy)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal dead letter policy for subscription %s: %w", s.GetName(), err)
		}
		data.DeadLetterPolicyJSON = policyJSON
	}

	// Convert retry policy to JSON
	if s.RetryPolicy != nil {
		policyJSON, err := json.Marshal(s.RetryPolicy)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal retry policy for subscription %s: %w", s.GetName(), err)
		}
		data.RetryPolicyJSON = policyJSON
	}

	return data, nil
}
