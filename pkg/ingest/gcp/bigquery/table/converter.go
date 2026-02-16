package table

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
)

// TableData holds converted BigQuery table data ready for Ent insertion.
type TableData struct {
	ID                          string
	DatasetID                   string
	FriendlyName                string
	Description                 string
	SchemaJSON                  json.RawMessage
	NumBytes                    *int64
	NumLongTermBytes            *int64
	NumRows                     *uint64
	CreationTime                string
	ExpirationTime              string
	LastModifiedTime            string
	TableType                   string
	LabelsJSON                  json.RawMessage
	EncryptionConfigurationJSON json.RawMessage
	TimePartitioningJSON        json.RawMessage
	RangePartitioningJSON       json.RawMessage
	ClusteringJSON              json.RawMessage
	RequirePartitionFilter      bool
	Etag                        string
	ProjectID                   string
	CollectedAt                 time.Time
}

// ConvertTable converts a raw BigQuery table to Ent-compatible data.
func ConvertTable(raw TableRaw, projectID string, collectedAt time.Time) (*TableData, error) {
	meta := raw.Metadata

	datasetResourceName := fmt.Sprintf("projects/%s/datasets/%s", projectID, raw.DatasetID)
	data := &TableData{
		ID:                     fmt.Sprintf("projects/%s/datasets/%s/tables/%s", projectID, raw.DatasetID, raw.TableID),
		DatasetID:              datasetResourceName,
		FriendlyName:           meta.Name,
		Description:            meta.Description,
		RequirePartitionFilter: meta.RequirePartitionFilter,
		Etag:                   meta.ETag,
		ProjectID:              projectID,
		CollectedAt:            collectedAt,
	}

	// Table type
	switch meta.Type {
	case bigquery.RegularTable:
		data.TableType = "TABLE"
	case bigquery.ViewTable:
		data.TableType = "VIEW"
	case bigquery.ExternalTable:
		data.TableType = "EXTERNAL"
	case bigquery.MaterializedView:
		data.TableType = "MATERIALIZED_VIEW"
	case bigquery.Snapshot:
		data.TableType = "SNAPSHOT"
	default:
		data.TableType = fmt.Sprintf("%v", meta.Type)
	}

	// Numeric fields
	if meta.NumBytes > 0 {
		data.NumBytes = &meta.NumBytes
	}
	if meta.NumLongTermBytes > 0 {
		data.NumLongTermBytes = &meta.NumLongTermBytes
	}
	if meta.NumRows > 0 {
		data.NumRows = &meta.NumRows
	}

	// Time fields
	if !meta.CreationTime.IsZero() {
		data.CreationTime = meta.CreationTime.Format(time.RFC3339)
	}
	if !meta.ExpirationTime.IsZero() {
		data.ExpirationTime = meta.ExpirationTime.Format(time.RFC3339)
	}
	if !meta.LastModifiedTime.IsZero() {
		data.LastModifiedTime = meta.LastModifiedTime.Format(time.RFC3339)
	}

	// Schema to JSON
	if meta.Schema != nil {
		schemaJSON, err := json.Marshal(meta.Schema)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal schema for table %s: %w", raw.TableID, err)
		}
		data.SchemaJSON = schemaJSON
	}

	// Labels to JSON
	if len(meta.Labels) > 0 {
		labelsJSON, err := json.Marshal(meta.Labels)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal labels for table %s: %w", raw.TableID, err)
		}
		data.LabelsJSON = labelsJSON
	}

	// Encryption config to JSON
	if meta.EncryptionConfig != nil {
		encJSON, err := json.Marshal(meta.EncryptionConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal encryption config for table %s: %w", raw.TableID, err)
		}
		data.EncryptionConfigurationJSON = encJSON
	}

	// Time partitioning to JSON
	if meta.TimePartitioning != nil {
		tpJSON, err := json.Marshal(meta.TimePartitioning)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal time partitioning for table %s: %w", raw.TableID, err)
		}
		data.TimePartitioningJSON = tpJSON
	}

	// Range partitioning to JSON
	if meta.RangePartitioning != nil {
		rpJSON, err := json.Marshal(meta.RangePartitioning)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal range partitioning for table %s: %w", raw.TableID, err)
		}
		data.RangePartitioningJSON = rpJSON
	}

	// Clustering to JSON
	if meta.Clustering != nil {
		clJSON, err := json.Marshal(meta.Clustering)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal clustering for table %s: %w", raw.TableID, err)
		}
		data.ClusteringJSON = clJSON
	}

	return data, nil
}
