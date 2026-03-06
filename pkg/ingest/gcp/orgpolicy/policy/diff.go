package policy

import (
	entorgpolicy "danny.vn/hotpot/pkg/storage/ent/gcp/orgpolicy"
)

type PolicyDiff struct {
	IsNew     bool
	IsChanged bool
}

func (d *PolicyDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

func DiffPolicyData(old *entorgpolicy.BronzeGCPOrgPolicyPolicy, new *PolicyData) *PolicyDiff {
	diff := &PolicyDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Etag != new.Etag {
		diff.IsChanged = true
	}

	return diff
}
