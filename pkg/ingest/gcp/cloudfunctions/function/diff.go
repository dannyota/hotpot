package function

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// FunctionDiff represents changes between old and new Cloud Function state.
type FunctionDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *FunctionDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffFunctionData compares existing Ent entity with new FunctionData and returns differences.
func DiffFunctionData(old *ent.BronzeGCPCloudFunctionsFunction, new *FunctionData) *FunctionDiff {
	diff := &FunctionDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		old.Description != new.Description ||
		old.Environment != new.Environment ||
		old.State != new.State ||
		old.UpdateTime != new.UpdateTime ||
		old.CreateTime != new.CreateTime ||
		old.KmsKeyName != new.KmsKeyName ||
		old.URL != new.URL ||
		old.SatisfiesPzs != new.SatisfiesPzs ||
		old.Location != new.Location ||
		!bytes.Equal(old.BuildConfigJSON, new.BuildConfigJSON) ||
		!bytes.Equal(old.ServiceConfigJSON, new.ServiceConfigJSON) ||
		!bytes.Equal(old.EventTriggerJSON, new.EventTriggerJSON) ||
		!bytes.Equal(old.StateMessagesJSON, new.StateMessagesJSON) ||
		!bytes.Equal(old.LabelsJSON, new.LabelsJSON) {
		diff.IsChanged = true
	}

	return diff
}
