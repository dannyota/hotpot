package constraint

import (
	entorgpolicy "github.com/dannyota/hotpot/pkg/storage/ent/gcp/orgpolicy"
)

type ConstraintDiff struct {
	IsNew     bool
	IsChanged bool
}

func (d *ConstraintDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

func DiffConstraintData(old *entorgpolicy.BronzeGCPOrgPolicyConstraint, new *ConstraintData) *ConstraintDiff {
	diff := &ConstraintDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.DisplayName != new.DisplayName ||
		old.Description != new.Description ||
		old.ConstraintDefault != new.ConstraintDefault ||
		old.SupportsDryRun != new.SupportsDryRun ||
		old.SupportsSimulation != new.SupportsSimulation {
		diff.IsChanged = true
	}

	return diff
}
