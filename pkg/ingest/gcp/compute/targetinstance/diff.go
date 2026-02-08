package targetinstance

import (
	"hotpot/pkg/storage/ent"
)

// TargetInstanceDiff represents changes between old and new target instance states.
type TargetInstanceDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffTargetInstanceData compares existing Ent entity with new TargetInstanceData and returns differences.
func DiffTargetInstanceData(old *ent.BronzeGCPComputeTargetInstance, new *TargetInstanceData) *TargetInstanceDiff {
	diff := &TargetInstanceDiff{}

	// New target instance
	if old == nil {
		diff.IsNew = true
		return diff
	}

	// Compare fields
	if old.Name != new.Name ||
		old.Description != new.Description ||
		old.Zone != new.Zone ||
		old.Instance != new.Instance ||
		old.Network != new.Network ||
		old.NatPolicy != new.NatPolicy ||
		old.SecurityPolicy != new.SecurityPolicy ||
		old.SelfLink != new.SelfLink ||
		old.CreationTimestamp != new.CreationTimestamp {
		diff.IsChanged = true
	}

	return diff
}

// HasAnyChange returns true if any part of the target instance changed.
func (d *TargetInstanceDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
