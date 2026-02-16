package secret

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// SecretData holds converted secret data ready for Ent insertion.
type SecretData struct {
	ID                  string
	Name                string
	CreateTime          string
	Etag                string
	ReplicationJSON     json.RawMessage
	RotationJSON        json.RawMessage
	TopicsJSON          json.RawMessage
	VersionAliasesJSON  json.RawMessage
	AnnotationsJSON     json.RawMessage
	Labels              []LabelData
	ProjectID           string
	CollectedAt         time.Time
}

// LabelData holds converted label data.
type LabelData struct {
	Key   string
	Value string
}

// ConvertSecret converts a GCP API Secret to Ent-compatible data.
func ConvertSecret(s *secretmanagerpb.Secret, projectID string, collectedAt time.Time) (*SecretData, error) {
	data := &SecretData{
		ID:          s.Name,
		Name:        s.Name,
		Etag:        s.Etag,
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}

	if s.CreateTime != nil {
		data.CreateTime = s.CreateTime.AsTime().Format(time.RFC3339)
	}

	if s.Replication != nil {
		j, err := json.Marshal(s.Replication)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal replication for secret %s: %w", s.Name, err)
		}
		data.ReplicationJSON = j
	}

	if s.Rotation != nil {
		j, err := json.Marshal(s.Rotation)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal rotation for secret %s: %w", s.Name, err)
		}
		data.RotationJSON = j
	}

	if len(s.Topics) > 0 {
		j, err := json.Marshal(s.Topics)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal topics for secret %s: %w", s.Name, err)
		}
		data.TopicsJSON = j
	}

	if len(s.VersionAliases) > 0 {
		j, err := json.Marshal(s.VersionAliases)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal version_aliases for secret %s: %w", s.Name, err)
		}
		data.VersionAliasesJSON = j
	}

	if len(s.Annotations) > 0 {
		j, err := json.Marshal(s.Annotations)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal annotations for secret %s: %w", s.Name, err)
		}
		data.AnnotationsJSON = j
	}

	// Convert labels map to child data
	data.Labels = ConvertLabels(s.Labels)

	return data, nil
}

// ConvertLabels converts a labels map to LabelData slice.
func ConvertLabels(labels map[string]string) []LabelData {
	if len(labels) == 0 {
		return nil
	}

	result := make([]LabelData, 0, len(labels))
	for k, v := range labels {
		result = append(result, LabelData{Key: k, Value: v})
	}
	return result
}
