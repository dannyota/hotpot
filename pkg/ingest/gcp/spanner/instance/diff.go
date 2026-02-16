package instance

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// InstanceDiff represents changes between old and new Spanner instance state.
type InstanceDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *InstanceDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffInstanceData compares existing Ent entity with new InstanceData and returns differences.
func DiffInstanceData(old *ent.BronzeGCPSpannerInstance, new *InstanceData) *InstanceDiff {
	diff := &InstanceDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		old.Config != new.Config ||
		old.DisplayName != new.DisplayName ||
		old.NodeCount != new.NodeCount ||
		old.ProcessingUnits != new.ProcessingUnits ||
		old.State != new.State ||
		old.CreateTime != new.CreateTime ||
		old.UpdateTime != new.UpdateTime ||
		old.Edition != new.Edition ||
		old.DefaultBackupScheduleType != new.DefaultBackupScheduleType ||
		!bytes.Equal(old.LabelsJSON, new.LabelsJSON) ||
		!bytes.Equal(old.EndpointUrisJSON, new.EndpointUrisJSON) {
		diff.IsChanged = true
	}

	return diff
}
