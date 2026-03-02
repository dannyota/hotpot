package installed_software

import (
	entinventory "github.com/dannyota/hotpot/pkg/storage/ent/meec/inventory"
)

// InstalledSoftwareDiff represents changes between old and new installed software states.
type InstalledSoftwareDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffInstalledSoftwareData compares old Ent entity and new data.
func DiffInstalledSoftwareData(old *entinventory.BronzeMEECInventoryInstalledSoftware, new *InstalledSoftwareData) *InstalledSoftwareDiff {
	if old == nil {
		return &InstalledSoftwareDiff{IsNew: true}
	}

	changed := old.ComputerResourceID != new.ComputerResourceID ||
		old.SoftwareID != new.SoftwareID ||
		old.SoftwareName != new.SoftwareName ||
		old.SoftwareVersion != new.SoftwareVersion ||
		old.DisplayName != new.DisplayName ||
		old.ManufacturerName != new.ManufacturerName ||
		old.InstalledDate != new.InstalledDate ||
		old.Architecture != new.Architecture ||
		old.Location != new.Location ||
		old.SwType != new.SwType ||
		old.SwCategoryName != new.SwCategoryName ||
		old.DetectedTime != new.DetectedTime

	return &InstalledSoftwareDiff{IsChanged: changed}
}
