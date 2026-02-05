package project

import (
	"hotpot/pkg/base/models/bronze"
)

// ProjectDiff represents changes between old and new project state.
type ProjectDiff struct {
	IsNew      bool
	IsChanged  bool
	LabelsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	HasChanges bool
}

// HasAnyChange returns true if there are any changes.
func (d *ProjectDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.LabelsDiff.HasChanges
}

// DiffProject compares old and new project states and returns differences.
func DiffProject(old, new *bronze.GCPProject) *ProjectDiff {
	diff := &ProjectDiff{}

	// New project
	if old == nil {
		diff.IsNew = true
		return diff
	}

	// Compare core fields
	if old.DisplayName != new.DisplayName ||
		old.State != new.State ||
		old.Parent != new.Parent ||
		old.UpdateTime != new.UpdateTime ||
		old.DeleteTime != new.DeleteTime ||
		old.Etag != new.Etag {
		diff.IsChanged = true
	}

	// Compare labels
	diff.LabelsDiff = diffLabels(old.Labels, new.Labels)

	return diff
}

// diffLabels compares two sets of labels.
func diffLabels(old, new []bronze.GCPProjectLabel) ChildDiff {
	diff := ChildDiff{}

	if len(old) != len(new) {
		diff.HasChanges = true
		return diff
	}

	// Build map of old labels
	oldMap := make(map[string]string, len(old))
	for _, l := range old {
		oldMap[l.Key] = l.Value
	}

	// Compare with new labels
	for _, l := range new {
		if oldValue, ok := oldMap[l.Key]; !ok || oldValue != l.Value {
			diff.HasChanges = true
			return diff
		}
	}

	return diff
}
