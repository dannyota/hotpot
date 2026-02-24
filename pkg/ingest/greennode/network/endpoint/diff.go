package endpoint

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// EndpointDiff represents changes between old and new endpoint states.
type EndpointDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffEndpointData compares old Ent entity and new EndpointData.
func DiffEndpointData(old *ent.BronzeGreenNodeNetworkEndpoint, new *EndpointData) *EndpointDiff {
	if old == nil {
		return &EndpointDiff{IsNew: true}
	}

	return &EndpointDiff{
		IsChanged: old.Name != new.Name ||
			old.Ipv4Address != new.Ipv4Address ||
			old.EndpointURL != new.EndpointURL ||
			old.Status != new.Status ||
			old.VpcID != new.VpcID,
	}
}

// HasAnyChange returns true if the endpoint changed.
func (d *EndpointDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
