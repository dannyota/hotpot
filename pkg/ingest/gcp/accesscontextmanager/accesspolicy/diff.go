package accesspolicy

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// AccessPolicyDiff represents changes between old and new access policy state.
type AccessPolicyDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *AccessPolicyDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffAccessPolicyData compares existing Ent entity with new AccessPolicyData and returns differences.
func DiffAccessPolicyData(old *ent.BronzeGCPAccessContextManagerAccessPolicy, new *AccessPolicyData) *AccessPolicyDiff {
	diff := &AccessPolicyDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Parent != new.Parent ||
		old.Title != new.Title ||
		old.Etag != new.Etag ||
		!bytes.Equal(old.ScopesJSON, new.ScopesJSON) {
		diff.IsChanged = true
	}

	return diff
}
