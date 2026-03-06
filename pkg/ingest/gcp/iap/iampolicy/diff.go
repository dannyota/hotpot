package iampolicy

import (
	"bytes"

	entiap "danny.vn/hotpot/pkg/storage/ent/gcp/iap"
)

// IAMPolicyDiff represents changes between old and new IAP IAM policy state.
type IAMPolicyDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *IAMPolicyDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffIAMPolicyData compares existing Ent entity with new IAMPolicyData and returns differences.
func DiffIAMPolicyData(old *entiap.BronzeGCPIAPIAMPolicy, new *IAMPolicyData) *IAMPolicyDiff {
	diff := &IAMPolicyDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		old.Etag != new.Etag ||
		old.Version != new.Version ||
		!bytes.Equal(old.BindingsJSON, new.BindingsJSON) ||
		!bytes.Equal(old.AuditConfigsJSON, new.AuditConfigsJSON) {
		diff.IsChanged = true
	}

	return diff
}
