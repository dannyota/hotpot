package volumetype

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// VolumeTypeDiff represents changes between old and new volume type states.
type VolumeTypeDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffVolumeTypeData compares old Ent entity and new VolumeTypeData.
func DiffVolumeTypeData(old *ent.BronzeGreenNodeVolumeVolumeType, new *VolumeTypeData) *VolumeTypeDiff {
	if old == nil {
		return &VolumeTypeDiff{IsNew: true}
	}

	return &VolumeTypeDiff{
		IsChanged: old.Name != new.Name ||
			old.Iops != new.Iops ||
			old.MaxSize != new.MaxSize ||
			old.MinSize != new.MinSize ||
			old.ThroughPut != new.ThroughPut ||
			old.ZoneID != new.ZoneID,
	}
}

// HasAnyChange returns true if the volume type changed.
func (d *VolumeTypeDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
