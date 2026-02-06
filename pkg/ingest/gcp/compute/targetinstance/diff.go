package targetinstance

import (
	"hotpot/pkg/base/models/bronze"
)

// TargetInstanceDiff represents changes between old and new target instance states.
type TargetInstanceDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffTargetInstance compares old and new target instance states.
func DiffTargetInstance(old, new *bronze.GCPComputeTargetInstance) *TargetInstanceDiff {
	if old == nil {
		return &TargetInstanceDiff{
			IsNew: true,
		}
	}

	return &TargetInstanceDiff{
		IsChanged: hasTargetInstanceFieldsChanged(old, new),
	}
}

// HasAnyChange returns true if any part of the target instance changed.
func (d *TargetInstanceDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// hasTargetInstanceFieldsChanged compares target instance fields.
func hasTargetInstanceFieldsChanged(old, new *bronze.GCPComputeTargetInstance) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Zone != new.Zone ||
		old.Instance != new.Instance ||
		old.Network != new.Network ||
		old.NatPolicy != new.NatPolicy ||
		old.SecurityPolicy != new.SecurityPolicy
}
