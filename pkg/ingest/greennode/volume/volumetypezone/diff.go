package volumetypezone

import (
	"bytes"

	entvol "danny.vn/hotpot/pkg/storage/ent/greennode/volume"
)

// VolumeTypeZoneDiff represents changes between old and new volume type zone states.
type VolumeTypeZoneDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffVolumeTypeZoneData compares old Ent entity and new VolumeTypeZoneData.
func DiffVolumeTypeZoneData(old *entvol.BronzeGreenNodeVolumeVolumeTypeZone, new *VolumeTypeZoneData) *VolumeTypeZoneDiff {
	if old == nil {
		return &VolumeTypeZoneDiff{IsNew: true}
	}

	return &VolumeTypeZoneDiff{
		IsChanged: old.Name != new.Name ||
			!bytes.Equal(old.PoolNameJSON, new.PoolNameJSON),
	}
}

// HasAnyChange returns true if the volume type zone changed.
func (d *VolumeTypeZoneDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
