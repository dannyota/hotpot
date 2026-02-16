package customconstraint

import (
	"fmt"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

type CustomConstraintDiff struct {
	IsNew     bool
	IsChanged bool
}

func (d *CustomConstraintDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

func DiffCustomConstraintData(old *ent.BronzeGCPOrgPolicyCustomConstraint, new *CustomConstraintData) *CustomConstraintDiff {
	diff := &CustomConstraintDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.DisplayName != new.DisplayName ||
		old.Description != new.Description ||
		old.Condition != new.Condition ||
		old.ActionType != new.ActionType ||
		fmt.Sprintf("%v", old.ResourceTypes) != fmt.Sprintf("%v", new.ResourceTypes) ||
		fmt.Sprintf("%v", old.MethodTypes) != fmt.Sprintf("%v", new.MethodTypes) {
		diff.IsChanged = true
	}

	return diff
}
