package peering

import (
	entnet "github.com/dannyota/hotpot/pkg/storage/ent/greennode/network"
)

// PeeringDiff represents changes between old and new peering states.
type PeeringDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffPeeringData compares old Ent entity and new PeeringData.
func DiffPeeringData(old *entnet.BronzeGreenNodeNetworkPeering, new *PeeringData) *PeeringDiff {
	if old == nil {
		return &PeeringDiff{IsNew: true}
	}

	return &PeeringDiff{
		IsChanged: old.Name != new.Name ||
			old.Status != new.Status ||
			old.FromVpcID != new.FromVpcID ||
			old.FromCidr != new.FromCidr ||
			old.EndVpcID != new.EndVpcID ||
			old.EndCidr != new.EndCidr,
	}
}

// HasAnyChange returns true if the peering changed.
func (d *PeeringDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
