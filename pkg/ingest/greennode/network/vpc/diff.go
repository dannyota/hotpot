package vpc

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// VPCDiff represents changes between old and new VPC states.
type VPCDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffVPCData compares old Ent entity and new VPCData.
func DiffVPCData(old *ent.BronzeGreenNodeNetworkVpc, new *VPCData) *VPCDiff {
	if old == nil {
		return &VPCDiff{IsNew: true}
	}

	return &VPCDiff{
		IsChanged: old.Name != new.Name ||
			old.Cidr != new.Cidr ||
			old.Status != new.Status ||
			old.RouteTableID != new.RouteTableID ||
			old.DNSStatus != new.DnsStatus ||
			old.DNSID != new.DnsID ||
			old.ZoneUUID != new.ZoneUuid,
	}
}

// HasAnyChange returns true if the VPC changed.
func (d *VPCDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
