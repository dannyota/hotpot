package vpc

import (
	entdo "github.com/dannyota/hotpot/pkg/storage/ent/do"
)

// VpcDiff represents changes between old and new VPC states.
type VpcDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffVpcData compares old Ent entity and new data.
func DiffVpcData(old *entdo.BronzeDOVpc, new *VpcData) *VpcDiff {
	if old == nil {
		return &VpcDiff{IsNew: true}
	}

	changed := old.Name != new.Name ||
		old.Description != new.Description ||
		old.Region != new.Region ||
		old.IPRange != new.IPRange ||
		old.Urn != new.URN ||
		old.IsDefault != new.IsDefault

	return &VpcDiff{IsChanged: changed}
}
