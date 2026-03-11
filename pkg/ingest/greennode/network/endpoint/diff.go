package endpoint

import (
	entnet "danny.vn/hotpot/pkg/storage/ent/greennode/network"
)

// EndpointDiff represents changes between old and new endpoint states.
type EndpointDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffEndpointData compares old Ent entity and new EndpointData.
func DiffEndpointData(old *entnet.BronzeGreenNodeNetworkEndpoint, new *EndpointData) *EndpointDiff {
	if old == nil {
		return &EndpointDiff{IsNew: true}
	}

	return &EndpointDiff{
		IsChanged: old.Name != new.Name ||
			old.Ipv4Address != new.Ipv4Address ||
			old.EndpointURL != new.EndpointURL ||
			old.EndpointAuthURL != new.EndpointAuthURL ||
			old.EndpointServiceID != new.EndpointServiceID ||
			old.Status != new.Status ||
			old.BillingStatus != new.BillingStatus ||
			old.EndpointType != new.EndpointType ||
			old.Version != new.Version ||
			old.Description != new.Description ||
			old.CreatedAt != new.CreatedAt ||
			old.UpdatedAt != new.UpdatedAt ||
			old.VpcID != new.VpcID ||
			old.VpcName != new.VpcName ||
			old.ZoneUUID != new.ZoneUuid ||
			old.EnableDNSName != new.EnableDnsName ||
			old.SubnetID != new.SubnetID ||
			old.CategoryName != new.CategoryName ||
			old.ServiceName != new.ServiceName ||
			old.ServiceEndpointType != new.ServiceEndpointType ||
			old.PackageName != new.PackageName,
	}
}

// HasAnyChange returns true if the endpoint changed.
func (d *EndpointDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
