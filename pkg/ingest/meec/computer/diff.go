package computer

import (
	entinventory "danny.vn/hotpot/pkg/storage/ent/meec/inventory"
)

// ComputerDiff represents changes between old and new computer states.
type ComputerDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if any part of the computer changed.
func (d *ComputerDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffComputerData compares old Ent entity and new data.
func DiffComputerData(old *entinventory.BronzeMEECInventoryComputer, new *ComputerData) *ComputerDiff {
	if old == nil {
		return &ComputerDiff{
			IsNew: true,
		}
	}

	diff := &ComputerDiff{}
	diff.IsChanged = hasComputerFieldsChanged(old, new)

	return diff
}

func hasComputerFieldsChanged(old *entinventory.BronzeMEECInventoryComputer, new *ComputerData) bool {
	return old.ResourceName != new.ResourceName ||
		old.FqdnName != new.FQDNName ||
		old.DomainNetbiosName != new.DomainNetbiosName ||
		old.IPAddress != new.IPAddress ||
		old.MACAddress != new.MACAddress ||
		old.OsName != new.OsName ||
		old.OsPlatform != new.OsPlatform ||
		old.OsPlatformName != new.OsPlatformName ||
		old.OsVersion != new.OsVersion ||
		old.ServicePack != new.ServicePack ||
		old.AgentVersion != new.AgentVersion ||
		old.ComputerLiveStatus != new.ComputerLiveStatus ||
		old.InstallationStatus != new.InstallationStatus ||
		old.ManagedStatus != new.ManagedStatus ||
		old.BranchOfficeName != new.BranchOfficeName ||
		old.Owner != new.Owner ||
		old.OwnerEmailID != new.OwnerEmailID ||
		old.Description != new.Description ||
		old.Location != new.Location ||
		old.LastSyncTime != new.LastSyncTime ||
		old.AgentLastContactTime != new.AgentLastContactTime ||
		old.AgentInstalledOn != new.AgentInstalledOn ||
		old.CustomerName != new.CustomerName ||
		old.CustomerID != new.CustomerID
}
