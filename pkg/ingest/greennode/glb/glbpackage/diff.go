package glbpackage

import (
	"bytes"

	entglb "danny.vn/hotpot/pkg/storage/ent/greennode/glb"
)

// GLBPackageDiff represents changes between old and new package states.
type GLBPackageDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffGLBPackageData compares old Ent entity and new GLBPackageData.
func DiffGLBPackageData(old *entglb.BronzeGreenNodeGLBGlobalPackage, new *GLBPackageData) *GLBPackageDiff {
	if old == nil {
		return &GLBPackageDiff{IsNew: true}
	}

	return &GLBPackageDiff{
		IsChanged: old.Name != new.Name ||
			old.Description != new.Description ||
			old.DescriptionEn != new.DescriptionEn ||
			!bytes.Equal(old.DetailJSON, new.DetailJSON) ||
			old.Enabled != new.Enabled ||
			old.BaseSku != new.BaseSku ||
			old.BaseConnectionRate != new.BaseConnectionRate ||
			old.BaseDomesticTrafficTotal != new.BaseDomesticTrafficTotal ||
			old.BaseNonDomesticTrafficTotal != new.BaseNonDomesticTrafficTotal ||
			old.ConnectionSku != new.ConnectionSku ||
			old.DomesticTrafficSku != new.DomesticTrafficSku ||
			old.NonDomesticTrafficSku != new.NonDomesticTrafficSku ||
			old.CreatedAtAPI != new.CreatedAtAPI ||
			old.UpdatedAtAPI != new.UpdatedAtAPI ||
			!bytes.Equal(old.VlbPackagesJSON, new.VlbPackagesJSON),
	}
}

// HasAnyChange returns true if the package changed.
func (d *GLBPackageDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
