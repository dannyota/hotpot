package glbregion

import (
	entglb "github.com/dannyota/hotpot/pkg/storage/ent/greennode/glb"
)

// GLBRegionDiff represents changes between old and new region states.
type GLBRegionDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffGLBRegionData compares old Ent entity and new GLBRegionData.
func DiffGLBRegionData(old *entglb.BronzeGreenNodeGLBGlobalRegion, new *GLBRegionData) *GLBRegionDiff {
	if old == nil {
		return &GLBRegionDiff{IsNew: true}
	}

	return &GLBRegionDiff{
		IsChanged: old.Name != new.Name ||
			old.VserverEndpoint != new.VserverEndpoint ||
			old.VlbEndpoint != new.VlbEndpoint ||
			old.UIServerEndpoint != new.UIServerEndpoint,
	}
}

// HasAnyChange returns true if the region changed.
func (d *GLBRegionDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
