package folder

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// FolderDiff represents changes between old and new folder state.
type FolderDiff struct {
	IsNew      bool
	IsChanged  bool
	LabelsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	HasChanges bool
}

// HasAnyChange returns true if there are any changes.
func (d *FolderDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.LabelsDiff.HasChanges
}

// DiffFolderData compares existing Ent entity with new FolderData and returns differences.
func DiffFolderData(old *ent.BronzeGCPFolder, new *FolderData) *FolderDiff {
	diff := &FolderDiff{}

	// New folder
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
	diff.LabelsDiff = diffLabelsData(old.Edges.Labels, new.Labels)

	return diff
}

// diffLabelsData compares Ent labels with new label data.
func diffLabelsData(old []*ent.BronzeGCPFolderLabel, new []LabelData) ChildDiff {
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
