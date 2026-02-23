package region

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// RegionDiff represents changes between old and new region states.
type RegionDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffRegionData compares old Ent entity and new RegionData.
func DiffRegionData(old *ent.BronzeGreenNodePortalRegion, new *RegionData) *RegionDiff {
	if old == nil {
		return &RegionDiff{IsNew: true}
	}

	return &RegionDiff{
		IsChanged: old.Name != new.Name ||
			old.Description != new.Description,
	}
}

// HasAnyChange returns true if the region changed.
func (d *RegionDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
