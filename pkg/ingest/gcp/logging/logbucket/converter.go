package logbucket

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/logging/apiv2/loggingpb"
)

// LogBucketData holds converted log bucket data ready for Ent insertion.
type LogBucketData struct {
	ResourceID       string
	Name             string
	Description      string
	RetentionDays    int32
	Locked           bool
	LifecycleState   string
	AnalyticsEnabled bool
	ProjectID        string
	Location         string
	CmekSettingsJSON json.RawMessage
	IndexConfigsJSON json.RawMessage
	CollectedAt      time.Time
}

// ConvertLogBucket converts a GCP API LogBucket to Ent-compatible data.
func ConvertLogBucket(b *loggingpb.LogBucket, projectID string, collectedAt time.Time) (*LogBucketData, error) {
	data := &LogBucketData{
		ResourceID:       b.GetName(),
		Name:             b.GetName(),
		Description:      b.GetDescription(),
		RetentionDays:    b.GetRetentionDays(),
		Locked:           b.GetLocked(),
		LifecycleState:   b.GetLifecycleState().String(),
		AnalyticsEnabled: b.GetAnalyticsEnabled(),
		ProjectID:        projectID,
		CollectedAt:      collectedAt,
	}

	// Extract location from resource name: projects/{project}/locations/{location}/buckets/{bucket}
	data.Location = extractLocation(b.GetName())

	if b.GetCmekSettings() != nil {
		cmekJSON, err := json.Marshal(b.GetCmekSettings())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal cmek_settings for bucket %s: %w", b.GetName(), err)
		}
		data.CmekSettingsJSON = cmekJSON
	}

	if b.GetIndexConfigs() != nil {
		indexJSON, err := json.Marshal(b.GetIndexConfigs())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal index_configs for bucket %s: %w", b.GetName(), err)
		}
		data.IndexConfigsJSON = indexJSON
	}

	return data, nil
}

// extractLocation extracts the location from a bucket resource name.
// Format: projects/{project}/locations/{location}/buckets/{bucket}
func extractLocation(name string) string {
	// Simple parsing: find "locations/" and extract until next "/"
	const prefix = "locations/"
	idx := 0
	for i := 0; i <= len(name)-len(prefix); i++ {
		if name[i:i+len(prefix)] == prefix {
			idx = i + len(prefix)
			break
		}
	}
	if idx == 0 {
		return ""
	}
	end := idx
	for end < len(name) && name[end] != '/' {
		end++
	}
	return name[idx:end]
}
