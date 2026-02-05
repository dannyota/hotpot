package network

import (
	"hotpot/pkg/base/models/bronze"
)

// NetworkDiff represents changes between old and new network states.
type NetworkDiff struct {
	IsNew     bool
	IsChanged bool

	// Child diffs (for granular tracking)
	PeeringsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffNetwork compares old and new network states.
// Returns nil if old is nil (new network).
func DiffNetwork(old, new *bronze.GCPComputeNetwork) *NetworkDiff {
	if old == nil {
		return &NetworkDiff{
			IsNew:        true,
			PeeringsDiff: ChildDiff{Changed: true},
		}
	}

	diff := &NetworkDiff{}

	// Compare network-level fields
	diff.IsChanged = hasNetworkFieldsChanged(old, new)

	// Compare children
	diff.PeeringsDiff = diffPeerings(old.Peerings, new.Peerings)

	return diff
}

// HasAnyChange returns true if any part of the network changed.
func (d *NetworkDiff) HasAnyChange() bool {
	if d.IsNew || d.IsChanged {
		return true
	}
	return d.PeeringsDiff.Changed
}

// hasNetworkFieldsChanged compares network-level fields (excluding children).
func hasNetworkFieldsChanged(old, new *bronze.GCPComputeNetwork) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.AutoCreateSubnetworks != new.AutoCreateSubnetworks ||
		old.Mtu != new.Mtu ||
		old.RoutingMode != new.RoutingMode ||
		old.NetworkFirewallPolicyEnforcementOrder != new.NetworkFirewallPolicyEnforcementOrder ||
		old.EnableUlaInternalIpv6 != new.EnableUlaInternalIpv6 ||
		old.InternalIpv6Range != new.InternalIpv6Range ||
		old.GatewayIpv4 != new.GatewayIpv4 ||
		old.SubnetworksJSON != new.SubnetworksJSON
}

func diffPeerings(old, new []bronze.GCPComputeNetworkPeering) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}

	// Build map of old peerings by name
	oldMap := make(map[string]bronze.GCPComputeNetworkPeering)
	for _, p := range old {
		oldMap[p.Name] = p
	}

	// Compare each new peering
	for _, newP := range new {
		oldP, ok := oldMap[newP.Name]
		if !ok {
			return ChildDiff{Changed: true}
		}
		if peeringChanged(oldP, newP) {
			return ChildDiff{Changed: true}
		}
	}

	return ChildDiff{Changed: false}
}

func peeringChanged(old, new bronze.GCPComputeNetworkPeering) bool {
	return old.Network != new.Network ||
		old.State != new.State ||
		old.StateDetails != new.StateDetails ||
		old.ExportCustomRoutes != new.ExportCustomRoutes ||
		old.ImportCustomRoutes != new.ImportCustomRoutes ||
		old.ExportSubnetRoutesWithPublicIp != new.ExportSubnetRoutesWithPublicIp ||
		old.ImportSubnetRoutesWithPublicIp != new.ImportSubnetRoutesWithPublicIp ||
		old.ExchangeSubnetRoutes != new.ExchangeSubnetRoutes ||
		old.StackType != new.StackType ||
		old.PeerMtu != new.PeerMtu ||
		old.AutoCreateRoutes != new.AutoCreateRoutes
}
