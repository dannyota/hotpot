package vpngateway

import (
	"hotpot/pkg/base/jsonb"
	"hotpot/pkg/base/models/bronze"
)

// VpnGatewayDiff represents changes between old and new VPN gateway states.
type VpnGatewayDiff struct {
	IsNew     bool
	IsChanged bool

	LabelsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffVpnGateway compares old and new VPN gateway states.
func DiffVpnGateway(old, new *bronze.GCPComputeVpnGateway) *VpnGatewayDiff {
	if old == nil {
		return &VpnGatewayDiff{
			IsNew:      true,
			LabelsDiff: ChildDiff{Changed: true},
		}
	}

	diff := &VpnGatewayDiff{}
	diff.IsChanged = hasVpnGatewayFieldsChanged(old, new)
	diff.LabelsDiff = diffLabels(old.Labels, new.Labels)

	return diff
}

// HasAnyChange returns true if any part of the VPN gateway changed.
func (d *VpnGatewayDiff) HasAnyChange() bool {
	if d.IsNew || d.IsChanged {
		return true
	}
	return d.LabelsDiff.Changed
}

func hasVpnGatewayFieldsChanged(old, new *bronze.GCPComputeVpnGateway) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Region != new.Region ||
		old.Network != new.Network ||
		old.SelfLink != new.SelfLink ||
		old.CreationTimestamp != new.CreationTimestamp ||
		old.LabelFingerprint != new.LabelFingerprint ||
		old.GatewayIpVersion != new.GatewayIpVersion ||
		old.StackType != new.StackType ||
		jsonb.Changed(old.VpnInterfacesJSON, new.VpnInterfacesJSON)
}

func diffLabels(old, new []bronze.GCPComputeVpnGatewayLabel) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}

	oldMap := make(map[string]string)
	for _, l := range old {
		oldMap[l.Key] = l.Value
	}
	for _, l := range new {
		if v, ok := oldMap[l.Key]; !ok || v != l.Value {
			return ChildDiff{Changed: true}
		}
	}

	return ChildDiff{Changed: false}
}
