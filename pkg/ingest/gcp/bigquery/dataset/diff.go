package dataset

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// DatasetDiff represents changes between old and new BigQuery dataset state.
type DatasetDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *DatasetDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffDatasetData compares existing Ent entity with new DatasetData and returns differences.
func DiffDatasetData(old *ent.BronzeGCPBigQueryDataset, new *DatasetData) *DatasetDiff {
	diff := &DatasetDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.FriendlyName != new.FriendlyName ||
		old.Description != new.Description ||
		old.Location != new.Location ||
		old.DefaultCollation != new.DefaultCollation ||
		old.Etag != new.Etag ||
		old.CreationTime != new.CreationTime ||
		old.LastModifiedTime != new.LastModifiedTime ||
		!nillableInt64Equal(old.DefaultTableExpirationMs, new.DefaultTableExpirationMs) ||
		!nillableInt64Equal(old.DefaultPartitionExpirationMs, new.DefaultPartitionExpirationMs) ||
		!nillableIntEqual(old.MaxTimeTravelHours, new.MaxTimeTravelHours) ||
		!bytes.Equal(old.LabelsJSON, new.LabelsJSON) ||
		!bytes.Equal(old.AccessJSON, new.AccessJSON) ||
		!bytes.Equal(old.DefaultEncryptionConfigurationJSON, new.DefaultEncryptionConfigurationJSON) {
		diff.IsChanged = true
	}

	return diff
}

func nillableInt64Equal(a, b *int64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func nillableIntEqual(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
