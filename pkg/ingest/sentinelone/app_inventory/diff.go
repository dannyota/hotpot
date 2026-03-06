package app_inventory

import ents1 "danny.vn/hotpot/pkg/storage/ent/s1"

// AppInventoryDiff represents changes between old and new app inventory states.
type AppInventoryDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffAppInventoryData compares old Ent entity and new data.
func DiffAppInventoryData(old *ents1.BronzeS1AppInventory, new *AppInventoryData) *AppInventoryDiff {
	if old == nil {
		return &AppInventoryDiff{IsNew: true}
	}

	changed := old.ApplicationName != new.ApplicationName ||
		old.ApplicationVendor != new.ApplicationVendor ||
		old.EndpointsCount != new.EndpointsCount ||
		old.ApplicationVersionsCount != new.ApplicationVersionsCount ||
		old.Estimate != new.Estimate

	return &AppInventoryDiff{IsChanged: changed}
}
