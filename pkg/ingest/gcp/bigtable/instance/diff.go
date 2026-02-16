package instance

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// InstanceDiff represents changes between old and new Bigtable instance state.
type InstanceDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *InstanceDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffInstanceData compares existing Ent entity with new InstanceData and returns differences.
func DiffInstanceData(old *ent.BronzeGCPBigtableInstance, new *InstanceData) *InstanceDiff {
	diff := &InstanceDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.DisplayName != new.DisplayName ||
		old.State != new.State ||
		old.InstanceType != new.InstanceType ||
		old.CreateTime != new.CreateTime ||
		!nilBoolEqual(old.SatisfiesPzs, new.SatisfiesPzs) ||
		!bytes.Equal(old.LabelsJSON, new.LabelsJSON) {
		diff.IsChanged = true
	}

	return diff
}

// nilBoolEqual compares two *bool values for equality.
func nilBoolEqual(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
