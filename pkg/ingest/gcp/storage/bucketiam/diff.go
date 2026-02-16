package bucketiam

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// BucketIamPolicyDiff represents changes between old and new bucket IAM policy state.
type BucketIamPolicyDiff struct {
	IsNew        bool
	IsChanged    bool
	BindingsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	HasChanges bool
}

// HasAnyChange returns true if there are any changes.
func (d *BucketIamPolicyDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.BindingsDiff.HasChanges
}

// DiffBucketIamPolicyData compares existing Ent entity with new BucketIamPolicyData and returns differences.
func DiffBucketIamPolicyData(old *ent.BronzeGCPStorageBucketIamPolicy, new *BucketIamPolicyData) *BucketIamPolicyDiff {
	diff := &BucketIamPolicyDiff{}

	// New policy
	if old == nil {
		diff.IsNew = true
		return diff
	}

	// Compare core fields
	if old.BucketName != new.BucketName ||
		old.Etag != new.Etag ||
		old.Version != new.Version {
		diff.IsChanged = true
	}

	// Compare bindings
	diff.BindingsDiff = diffBindingsData(old.Edges.Bindings, new.Bindings)

	return diff
}

// diffBindingsData compares Ent bindings with new binding data.
func diffBindingsData(old []*ent.BronzeGCPStorageBucketIamPolicyBinding, new []BindingData) ChildDiff {
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

// bytesEqualOrBothNil compares two byte slices, treating nil and empty as equal.
func bytesEqualOrBothNil(a, b []byte) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	return bytes.Equal(a, b)
}
