package appservice

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// ServiceDiff represents changes between old and new App Engine service state.
type ServiceDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *ServiceDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffServiceData compares existing Ent entity with new ServiceData and returns differences.
func DiffServiceData(old *ent.BronzeGCPAppEngineService, new *ServiceData) *ServiceDiff {
	diff := &ServiceDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		!bytes.Equal(old.SplitJSON, new.SplitJSON) ||
		!bytes.Equal(old.LabelsJSON, new.LabelsJSON) ||
		!bytes.Equal(old.NetworkSettingsJSON, new.NetworkSettingsJSON) {
		diff.IsChanged = true
	}

	return diff
}
