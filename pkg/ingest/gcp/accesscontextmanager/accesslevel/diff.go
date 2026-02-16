package accesslevel

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// AccessLevelDiff represents changes between old and new access level state.
type AccessLevelDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *AccessLevelDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffAccessLevelData compares existing Ent entity with new AccessLevelData and returns differences.
func DiffAccessLevelData(old *ent.BronzeGCPAccessContextManagerAccessLevel, new *AccessLevelData) *AccessLevelDiff {
	diff := &AccessLevelDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Title != new.Title ||
		old.Description != new.Description ||
		!bytes.Equal(old.BasicJSON, new.BasicJSON) ||
		!bytes.Equal(old.CustomJSON, new.CustomJSON) {
		diff.IsChanged = true
	}

	return diff
}
