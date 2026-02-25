package iampolicysearch

import (
	"bytes"

	entcloudasset "github.com/dannyota/hotpot/pkg/storage/ent/gcp/cloudasset"
)

// IAMPolicySearchDiff represents changes between old and new IAM policy search state.
type IAMPolicySearchDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *IAMPolicySearchDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffIAMPolicySearchData compares existing Ent entity with new IAMPolicySearchData and returns differences.
func DiffIAMPolicySearchData(old *entcloudasset.BronzeGCPCloudAssetIAMPolicySearch, new *IAMPolicySearchData) *IAMPolicySearchDiff {
	diff := &IAMPolicySearchDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.AssetType != new.AssetType ||
		old.Project != new.Project ||
		old.Organization != new.Organization ||
		!bytes.Equal(old.FoldersJSON, new.FoldersJSON) ||
		!bytes.Equal(old.PolicyJSON, new.PolicyJSON) ||
		!bytes.Equal(old.ExplanationJSON, new.ExplanationJSON) {
		diff.IsChanged = true
	}

	return diff
}
