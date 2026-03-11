package source

import (
	entsecuritycenter "danny.vn/hotpot/pkg/storage/ent/gcp/securitycenter"
)

// SourceDiff represents changes between old and new SCC source state.
type SourceDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *SourceDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffSourceData compares existing Ent entity with new SourceData and returns differences.
func DiffSourceData(old *entsecuritycenter.BronzeGCPSecurityCenterSource, new *SourceData) *SourceDiff {
	diff := &SourceDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.DisplayName != new.DisplayName ||
		old.Description != new.Description ||
		old.CanonicalName != new.CanonicalName {
		diff.IsChanged = true
	}

	return diff
}
