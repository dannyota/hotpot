package targetvpngateway

import (
	"hotpot/pkg/base/jsonb"
	"hotpot/pkg/base/models/bronze"
)

// TargetVpnGatewayDiff represents changes between old and new target VPN gateway states.
type TargetVpnGatewayDiff struct {
	IsNew     bool
	IsChanged bool

	LabelsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffTargetVpnGateway compares old and new target VPN gateway states.
func DiffTargetVpnGateway(old, new *bronze.GCPComputeTargetVpnGateway) *TargetVpnGatewayDiff {
	if old == nil {
		return &TargetVpnGatewayDiff{
			IsNew:      true,
			LabelsDiff: ChildDiff{Changed: true},
		}
	}

	diff := &TargetVpnGatewayDiff{}
	diff.IsChanged = hasTargetVpnGatewayFieldsChanged(old, new)
	diff.LabelsDiff = diffLabels(old.Labels, new.Labels)

	return diff
}

// HasAnyChange returns true if any part of the target VPN gateway changed.
func (d *TargetVpnGatewayDiff) HasAnyChange() bool {
	if d.IsNew || d.IsChanged {
		return true
	}
	return d.LabelsDiff.Changed
}

func hasTargetVpnGatewayFieldsChanged(old, new *bronze.GCPComputeTargetVpnGateway) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Status != new.Status ||
		old.Region != new.Region ||
		old.Network != new.Network ||
		old.LabelFingerprint != new.LabelFingerprint ||
		jsonb.Changed(old.ForwardingRulesJSON, new.ForwardingRulesJSON) ||
		jsonb.Changed(old.TunnelsJSON, new.TunnelsJSON)
}

func diffLabels(old, new []bronze.GCPComputeTargetVpnGatewayLabel) ChildDiff {
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
