package resourcesearch

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// ResourceSearchDiff represents changes between old and new resource search state.
type ResourceSearchDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *ResourceSearchDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffResourceSearchData compares existing Ent entity with new ResourceSearchData and returns differences.
func DiffResourceSearchData(old *ent.BronzeGCPCloudAssetResourceSearch, new *ResourceSearchData) *ResourceSearchDiff {
	diff := &ResourceSearchDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.AssetType != new.AssetType ||
		old.Project != new.Project ||
		old.DisplayName != new.DisplayName ||
		old.Description != new.Description ||
		old.Location != new.Location ||
		!bytes.Equal(old.LabelsJSON, new.LabelsJSON) ||
		!bytes.Equal(old.NetworkTagsJSON, new.NetworkTagsJSON) ||
		!bytes.Equal(old.AdditionalAttributesJSON, new.AdditionalAttributesJSON) {
		diff.IsChanged = true
	}

	return diff
}
