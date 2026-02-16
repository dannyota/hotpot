package bucket

import (
	"encoding/json"
	"fmt"
	"time"

	storagev1 "google.golang.org/api/storage/v1"
)

// BucketData holds converted bucket data ready for Ent insertion.
type BucketData struct {
	ID                   string
	Name                 string
	Location             string
	StorageClass         string
	ProjectNumber        string
	TimeCreated          string
	Updated              string
	DefaultEventBasedHold bool
	Metageneration       string
	Etag                 string
	IamConfigurationJSON json.RawMessage
	EncryptionJSON       json.RawMessage
	LifecycleJSON        json.RawMessage
	VersioningJSON       json.RawMessage
	RetentionPolicyJSON  json.RawMessage
	LoggingJSON          json.RawMessage
	CorsJSON             json.RawMessage
	WebsiteJSON          json.RawMessage
	AutoclassJSON        json.RawMessage
	Labels               []LabelData
	ProjectID            string
	CollectedAt          time.Time
}

// LabelData holds converted label data.
type LabelData struct {
	Key   string
	Value string
}

// ConvertBucket converts a GCP API Bucket to Ent-compatible data.
func ConvertBucket(b *storagev1.Bucket, projectID string, collectedAt time.Time) (*BucketData, error) {
	data := &BucketData{
		ID:                    b.Id,
		Name:                  b.Name,
		Location:              b.Location,
		StorageClass:          b.StorageClass,
		ProjectNumber:         fmt.Sprintf("%d", b.ProjectNumber),
		TimeCreated:           b.TimeCreated,
		Updated:               b.Updated,
		DefaultEventBasedHold: b.DefaultEventBasedHold,
		Metageneration:        fmt.Sprintf("%d", b.Metageneration),
		Etag:                  b.Etag,
		ProjectID:             projectID,
		CollectedAt:           collectedAt,
	}

	if b.IamConfiguration != nil {
		j, err := json.Marshal(b.IamConfiguration)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal iam_configuration for bucket %s: %w", b.Name, err)
		}
		data.IamConfigurationJSON = j
	}

	if b.Encryption != nil {
		j, err := json.Marshal(b.Encryption)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal encryption for bucket %s: %w", b.Name, err)
		}
		data.EncryptionJSON = j
	}

	if b.Lifecycle != nil {
		j, err := json.Marshal(b.Lifecycle)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal lifecycle for bucket %s: %w", b.Name, err)
		}
		data.LifecycleJSON = j
	}

	if b.Versioning != nil {
		j, err := json.Marshal(b.Versioning)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal versioning for bucket %s: %w", b.Name, err)
		}
		data.VersioningJSON = j
	}

	if b.RetentionPolicy != nil {
		j, err := json.Marshal(b.RetentionPolicy)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal retention_policy for bucket %s: %w", b.Name, err)
		}
		data.RetentionPolicyJSON = j
	}

	if b.Logging != nil {
		j, err := json.Marshal(b.Logging)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal logging for bucket %s: %w", b.Name, err)
		}
		data.LoggingJSON = j
	}

	if b.Cors != nil {
		j, err := json.Marshal(b.Cors)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal cors for bucket %s: %w", b.Name, err)
		}
		data.CorsJSON = j
	}

	if b.Website != nil {
		j, err := json.Marshal(b.Website)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal website for bucket %s: %w", b.Name, err)
		}
		data.WebsiteJSON = j
	}

	if b.Autoclass != nil {
		j, err := json.Marshal(b.Autoclass)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal autoclass for bucket %s: %w", b.Name, err)
		}
		data.AutoclassJSON = j
	}

	// Convert labels map to child data
	data.Labels = ConvertLabels(b.Labels)

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
