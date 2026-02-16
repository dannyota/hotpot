package enabledservice

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// EnabledServiceDiff represents changes between old and new enabled service states.
type EnabledServiceDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *EnabledServiceDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffEnabledServiceData compares existing Ent entity with new EnabledServiceData and returns differences.
func DiffEnabledServiceData(old *ent.BronzeGCPServiceUsageEnabledService, new *EnabledServiceData) *EnabledServiceDiff {
	diff := &EnabledServiceDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		old.Parent != new.Parent ||
		old.State != new.State ||
		!bytes.Equal(old.ConfigJSON, new.ConfigJSON) {
		diff.IsChanged = true
	}

	return diff
}
