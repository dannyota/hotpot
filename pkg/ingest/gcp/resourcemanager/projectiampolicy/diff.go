package projectiampolicy

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// ProjectIamPolicyDiff represents changes between old and new project IAM policy state.
type ProjectIamPolicyDiff struct {
	IsNew        bool
	IsChanged    bool
	BindingsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	HasChanges bool
}

// HasAnyChange returns true if there are any changes.
func (d *ProjectIamPolicyDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.BindingsDiff.HasChanges
}

// DiffProjectIamPolicyData compares existing Ent entity with new ProjectIamPolicyData and returns differences.
func DiffProjectIamPolicyData(old *ent.BronzeGCPProjectIamPolicy, new *ProjectIamPolicyData) *ProjectIamPolicyDiff {
	diff := &ProjectIamPolicyDiff{}

	// New policy
	if old == nil {
		diff.IsNew = true
		return diff
	}

	// Compare core fields
	if old.ResourceName != new.ResourceName ||
		old.Etag != new.Etag ||
		old.Version != new.Version {
		diff.IsChanged = true
	}

	// Compare bindings
	diff.BindingsDiff = diffBindingsData(old.Edges.Bindings, new.Bindings)

	return diff
}

// diffBindingsData compares Ent bindings with new binding data.
func diffBindingsData(old []*ent.BronzeGCPProjectIamPolicyBinding, new []BindingData) ChildDiff {
	diff := ChildDiff{}

	if len(old) != len(new) {
		diff.HasChanges = true
		return diff
	}

	// Build map of old bindings by role
	type bindingKey struct {
		Role          string
		MembersJSON   string
		ConditionJSON string
	}

	oldSet := make(map[bindingKey]struct{}, len(old))
	for _, b := range old {
		key := bindingKey{
			Role:          b.Role,
			MembersJSON:   string(b.MembersJSON),
			ConditionJSON: string(b.ConditionJSON),
		}
		oldSet[key] = struct{}{}
	}

	for _, b := range new {
		key := bindingKey{
			Role: b.Role,
		}
		if b.MembersJSON != nil {
			key.MembersJSON = string(b.MembersJSON)
		}
		if b.ConditionJSON != nil {
			key.ConditionJSON = string(b.ConditionJSON)
		}

		if _, ok := oldSet[key]; !ok {
			diff.HasChanges = true
			return diff
		}
	}

	return diff
}
