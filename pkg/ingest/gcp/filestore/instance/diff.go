package instance

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// InstanceDiff represents changes between old and new Filestore instance state.
type InstanceDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *InstanceDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffInstanceData compares existing Ent entity with new InstanceData and returns differences.
func DiffInstanceData(old *ent.BronzeGCPFilestoreInstance, new *InstanceData) *InstanceDiff {
	if old == nil {
		return &InstanceDiff{IsNew: true}
	}

	diff := &InstanceDiff{}

	if old.Name != new.Name ||
		old.Description != new.Description ||
		old.State != new.State ||
		old.StatusMessage != new.StatusMessage ||
		old.CreateTime != new.CreateTime ||
		old.Tier != new.Tier ||
		old.Etag != new.Etag ||
		old.SatisfiesPzs != new.SatisfiesPzs ||
		old.SatisfiesPzi != new.SatisfiesPzi ||
		old.KmsKeyName != new.KmsKeyName ||
		old.MaxCapacityGB != new.MaxCapacityGB ||
		old.Protocol != new.Protocol ||
		old.Location != new.Location ||
		!bytes.Equal(old.LabelsJSON, new.LabelsJSON) ||
		!bytes.Equal(old.FileSharesJSON, new.FileSharesJSON) ||
		!bytes.Equal(old.NetworksJSON, new.NetworksJSON) ||
		!bytes.Equal(old.SuspensionReasonsJSON, new.SuspensionReasonsJSON) {
		diff.IsChanged = true
	}

	return diff
}
