package volume

import (
	"reflect"

	entdo "danny.vn/hotpot/pkg/storage/ent/do"
)

// VolumeDiff represents changes between old and new Volume states.
type VolumeDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffVolumeData compares old Ent entity and new data.
func DiffVolumeData(old *entdo.BronzeDOVolume, new *VolumeData) *VolumeDiff {
	if old == nil {
		return &VolumeDiff{IsNew: true}
	}

	changed := old.Name != new.Name ||
		old.Region != new.Region ||
		old.SizeGigabytes != new.SizeGigabytes ||
		old.Description != new.Description ||
		old.FilesystemType != new.FilesystemType ||
		old.FilesystemLabel != new.FilesystemLabel ||
		!reflect.DeepEqual(old.DropletIdsJSON, new.DropletIdsJSON) ||
		!reflect.DeepEqual(old.TagsJSON, new.TagsJSON)

	return &VolumeDiff{IsChanged: changed}
}
