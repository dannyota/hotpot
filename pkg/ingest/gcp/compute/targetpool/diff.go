package targetpool

import (
	"encoding/json"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// TargetPoolDiff represents changes between old and new target pool states.
type TargetPoolDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffTargetPoolData compares existing Ent entity with new TargetPoolData.
func DiffTargetPoolData(old *ent.BronzeGCPComputeTargetPool, new *TargetPoolData) *TargetPoolDiff {
	diff := &TargetPoolDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		old.Description != new.Description ||
		old.CreationTimestamp != new.CreationTimestamp ||
		old.SelfLink != new.SelfLink ||
		old.SessionAffinity != new.SessionAffinity ||
		old.BackupPool != new.BackupPool ||
		old.FailoverRatio != new.FailoverRatio ||
		old.SecurityPolicy != new.SecurityPolicy ||
		old.Region != new.Region ||
		jsonChanged(old.HealthChecksJSON, new.HealthChecksJSON) ||
		jsonChanged(old.InstancesJSON, new.InstancesJSON) {
		diff.IsChanged = true
	}

	return diff
}

func jsonChanged(old, new []interface{}) bool {
	oldBytes, _ := json.Marshal(old)
	newBytes, _ := json.Marshal(new)
	return string(oldBytes) != string(newBytes)
}

// HasAnyChange returns true if any part of the target pool changed.
func (d *TargetPoolDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
