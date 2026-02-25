package subnet

import (
	entnet "github.com/dannyota/hotpot/pkg/storage/ent/greennode/network"
)

// SubnetDiff represents changes between old and new subnet states.
type SubnetDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffSubnetData compares old Ent entity and new SubnetData.
func DiffSubnetData(old *entnet.BronzeGreenNodeNetworkSubnet, new *SubnetData) *SubnetDiff {
	if old == nil {
		return &SubnetDiff{IsNew: true}
	}

	return &SubnetDiff{
		IsChanged: old.Name != new.Name ||
			old.NetworkID != new.NetworkID ||
			old.Cidr != new.Cidr ||
			old.Status != new.Status ||
			old.RouteTableID != new.RouteTableID ||
			old.InterfaceACLPolicyID != new.InterfaceAclPolicyID ||
			old.ZoneID != new.ZoneID,
	}
}

// HasAnyChange returns true if the subnet changed.
func (d *SubnetDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
