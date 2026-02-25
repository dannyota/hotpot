package userimage

import (
	entcompute "github.com/dannyota/hotpot/pkg/storage/ent/greennode/compute"
)

// UserImageDiff represents changes between old and new user image states.
type UserImageDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffUserImageData compares old Ent entity and new UserImageData.
func DiffUserImageData(old *entcompute.BronzeGreenNodeComputeUserImage, new *UserImageData) *UserImageDiff {
	if old == nil {
		return &UserImageDiff{IsNew: true}
	}

	return &UserImageDiff{
		IsChanged: old.Name != new.Name ||
			old.Status != new.Status ||
			old.MinDisk != new.MinDisk ||
			old.ImageSize != new.ImageSize ||
			old.MetaData != new.MetaData,
	}
}

// HasAnyChange returns true if the user image changed.
func (d *UserImageDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
