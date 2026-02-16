package asset

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// AssetDiff represents changes between old and new Cloud Asset state.
type AssetDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *AssetDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffAssetData compares existing Ent entity with new AssetData and returns differences.
func DiffAssetData(old *ent.BronzeGCPCloudAssetAsset, new *AssetData) *AssetDiff {
	diff := &AssetDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.AssetType != new.AssetType ||
		old.UpdateTime != new.UpdateTime ||
		!bytes.Equal(old.ResourceJSON, new.ResourceJSON) ||
		!bytes.Equal(old.IamPolicyJSON, new.IamPolicyJSON) ||
		!bytes.Equal(old.OrgPolicyJSON, new.OrgPolicyJSON) ||
		!bytes.Equal(old.AccessPolicyJSON, new.AccessPolicyJSON) ||
		!bytes.Equal(old.OsInventoryJSON, new.OsInventoryJSON) {
		diff.IsChanged = true
	}

	return diff
}
