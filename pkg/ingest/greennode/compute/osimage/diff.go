package osimage

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// OSImageDiff represents changes between old and new OS image states.
type OSImageDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffOSImageData compares old Ent entity and new OSImageData.
func DiffOSImageData(old *ent.BronzeGreenNodeComputeOSImage, new *OSImageData) *OSImageDiff {
	if old == nil {
		return &OSImageDiff{IsNew: true}
	}

	return &OSImageDiff{
		IsChanged: old.ImageType != new.ImageType ||
			old.ImageVersion != new.ImageVersion ||
			!ptrBoolEqual(old.Licence, new.Licence) ||
			!ptrStringEqual(old.LicenseKey, new.LicenseKey) ||
			old.Description != new.Description ||
			old.ZoneID != new.ZoneID ||
			old.PackageLimitCPU != new.PackageLimitCpu ||
			old.PackageLimitMemory != new.PackageLimitMemory ||
			old.PackageLimitDiskSize != new.PackageLimitDiskSize,
	}
}

// HasAnyChange returns true if the OS image changed.
func (d *OSImageDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

func ptrBoolEqual(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func ptrStringEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
