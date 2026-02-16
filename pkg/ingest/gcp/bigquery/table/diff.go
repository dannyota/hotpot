package table

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// TableDiff represents changes between old and new BigQuery table state.
type TableDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *TableDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffTableData compares existing Ent entity with new TableData and returns differences.
func DiffTableData(old *ent.BronzeGCPBigQueryTable, new *TableData) *TableDiff {
	diff := &TableDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.DatasetID != new.DatasetID ||
		old.FriendlyName != new.FriendlyName ||
		old.Description != new.Description ||
		old.TableType != new.TableType ||
		old.RequirePartitionFilter != new.RequirePartitionFilter ||
		old.Etag != new.Etag ||
		old.CreationTime != new.CreationTime ||
		old.ExpirationTime != new.ExpirationTime ||
		old.LastModifiedTime != new.LastModifiedTime ||
		!nillableInt64Equal(old.NumBytes, new.NumBytes) ||
		!nillableInt64Equal(old.NumLongTermBytes, new.NumLongTermBytes) ||
		!nillableUint64Equal(old.NumRows, new.NumRows) ||
		!bytes.Equal(old.SchemaJSON, new.SchemaJSON) ||
		!bytes.Equal(old.LabelsJSON, new.LabelsJSON) ||
		!bytes.Equal(old.EncryptionConfigurationJSON, new.EncryptionConfigurationJSON) ||
		!bytes.Equal(old.TimePartitioningJSON, new.TimePartitioningJSON) ||
		!bytes.Equal(old.RangePartitioningJSON, new.RangePartitioningJSON) ||
		!bytes.Equal(old.ClusteringJSON, new.ClusteringJSON) {
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

func nillableUint64Equal(a, b *uint64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
