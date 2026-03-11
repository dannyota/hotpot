package dataset

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
)

// DatasetData holds converted BigQuery dataset data ready for Ent insertion.
type DatasetData struct {
	ID                                 string
	FriendlyName                       string
	Description                        string
	Location                           string
	DefaultTableExpirationMs           *int64
	DefaultPartitionExpirationMs       *int64
	LabelsJSON                         json.RawMessage
	AccessJSON                         json.RawMessage
	CreationTime                       string
	LastModifiedTime                   string
	Etag                               string
	DefaultCollation                   string
	MaxTimeTravelHours                 *int
	DefaultEncryptionConfigurationJSON json.RawMessage
	ProjectID                          string
	CollectedAt                        time.Time
}

// ConvertDataset converts a raw BigQuery dataset to Ent-compatible data.
func ConvertDataset(raw DatasetRaw, projectID string, collectedAt time.Time) (*DatasetData, error) {
	meta := raw.Metadata

	data := &DatasetData{
		ID:           fmt.Sprintf("projects/%s/datasets/%s", projectID, raw.DatasetID),
		FriendlyName: meta.Name,
		Description:  meta.Description,
		Location:     meta.Location,
		Etag:         meta.ETag,
		ProjectID:    projectID,
		CollectedAt:  collectedAt,
	}

	// Default table expiration
	if meta.DefaultTableExpiration > 0 {
		ms := meta.DefaultTableExpiration.Milliseconds()
		data.DefaultTableExpirationMs = &ms
	}

	// Default partition expiration
	if meta.DefaultPartitionExpiration > 0 {
		ms := meta.DefaultPartitionExpiration.Milliseconds()
		data.DefaultPartitionExpirationMs = &ms
	}

	// Creation time
	if !meta.CreationTime.IsZero() {
		data.CreationTime = meta.CreationTime.Format(time.RFC3339)
	}

	// Last modified time
	if !meta.LastModifiedTime.IsZero() {
		data.LastModifiedTime = meta.LastModifiedTime.Format(time.RFC3339)
	}

	// Max time travel hours
	if meta.MaxTimeTravel > 0 {
		hours := int(meta.MaxTimeTravel.Hours())
		data.MaxTimeTravelHours = &hours
	}

	// Labels to JSON
	if len(meta.Labels) > 0 {
		labelsJSON, err := json.Marshal(meta.Labels)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal labels for dataset %s: %w", raw.DatasetID, err)
		}
		data.LabelsJSON = labelsJSON
	}

	// Access entries to JSON
	if len(meta.Access) > 0 {
		accessJSON, err := json.Marshal(meta.Access)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal access for dataset %s: %w", raw.DatasetID, err)
		}
		data.AccessJSON = accessJSON
	}

	// Default encryption configuration to JSON
	if meta.DefaultEncryptionConfig != nil {
		encJSON, err := json.Marshal(meta.DefaultEncryptionConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal encryption config for dataset %s: %w", raw.DatasetID, err)
		}
		data.DefaultEncryptionConfigurationJSON = encJSON
	}

	return data, nil
}

// convertAccessEntries converts BigQuery access entries to a JSON-serializable form.
// This is kept as a helper in case we need custom conversion later.
func convertAccessEntries(entries []*bigquery.AccessEntry) []map[string]string {
	result := make([]map[string]string, 0, len(entries))
	for _, entry := range entries {
		m := map[string]string{
			"role":   string(entry.Role),
			"entity": entry.Entity,
		}
		if entry.EntityType != 0 {
			m["entityType"] = fmt.Sprintf("%d", entry.EntityType)
		}
		result = append(result, m)
	}
	return result
}
