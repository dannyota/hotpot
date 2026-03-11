package software

import entinventory "danny.vn/hotpot/pkg/storage/ent/meec/inventory"

// SoftwareDiff represents changes between old and new software states.
type SoftwareDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffSoftwareData compares old Ent entity and new data.
func DiffSoftwareData(old *entinventory.BronzeMEECInventorySoftware, new *SoftwareData) *SoftwareDiff {
	if old == nil {
		return &SoftwareDiff{IsNew: true}
	}

	changed := old.SoftwareName != new.SoftwareName ||
		old.SoftwareVersion != new.SoftwareVersion ||
		old.DisplayName != new.DisplayName ||
		old.ManufacturerID != new.ManufacturerID ||
		old.ManufacturerName != new.ManufacturerName ||
		old.SwCategoryName != new.SwCategoryName ||
		old.SwType != new.SwType ||
		old.SwFamily != new.SwFamily ||
		old.InstalledFormat != new.InstalledFormat ||
		old.IsUsageProhibited != new.IsUsageProhibited ||
		old.ManagedInstallations != new.ManagedInstallations ||
		old.NetworkInstallations != new.NetworkInstallations ||
		old.ManagedSwID != new.ManagedSwID ||
		old.DetectedTime != new.DetectedTime ||
		old.CompliantStatus != new.CompliantStatus ||
		old.TotalCopies != new.TotalCopies ||
		old.RemainingCopies != new.RemainingCopies

	return &SoftwareDiff{IsChanged: changed}
}
