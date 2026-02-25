package lbpackage

import (
	entlb "github.com/dannyota/hotpot/pkg/storage/ent/greennode/loadbalancer"
)

// PackageDiff represents changes between old and new package states.
type PackageDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffPackageData compares old Ent entity and new PackageData.
func DiffPackageData(old *entlb.BronzeGreenNodeLoadBalancerPackage, new *PackageData) *PackageDiff {
	if old == nil {
		return &PackageDiff{IsNew: true}
	}

	return &PackageDiff{
		IsChanged: old.Name != new.Name ||
			old.Type != new.Type ||
			old.ConnectionNumber != new.ConnectionNumber ||
			old.DataTransfer != new.DataTransfer ||
			old.Mode != new.Mode ||
			old.LbType != new.LbType ||
			old.DisplayLbType != new.DisplayLbType,
	}
}

// HasAnyChange returns true if the package changed.
func (d *PackageDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
