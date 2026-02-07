package vpntunnel

import (
	"hotpot/pkg/base/jsonb"
	"hotpot/pkg/base/models/bronze"
)

// VpnTunnelDiff represents changes between old and new VPN tunnel states.
type VpnTunnelDiff struct {
	IsNew     bool
	IsChanged bool

	LabelsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffVpnTunnel compares old and new VPN tunnel states.
func DiffVpnTunnel(old, new *bronze.GCPComputeVpnTunnel) *VpnTunnelDiff {
	if old == nil {
		return &VpnTunnelDiff{
			IsNew:      true,
			LabelsDiff: ChildDiff{Changed: true},
		}
	}

	diff := &VpnTunnelDiff{}
	diff.IsChanged = hasVpnTunnelFieldsChanged(old, new)
	diff.LabelsDiff = diffLabels(old.Labels, new.Labels)

	return diff
}

// HasAnyChange returns true if any part of the VPN tunnel changed.
func (d *VpnTunnelDiff) HasAnyChange() bool {
	if d.IsNew || d.IsChanged {
		return true
	}
	return d.LabelsDiff.Changed
}

func hasVpnTunnelFieldsChanged(old, new *bronze.GCPComputeVpnTunnel) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Status != new.Status ||
		old.DetailedStatus != new.DetailedStatus ||
		old.Region != new.Region ||
		old.LabelFingerprint != new.LabelFingerprint ||
		old.IkeVersion != new.IkeVersion ||
		old.PeerIp != new.PeerIp ||
		old.PeerExternalGateway != new.PeerExternalGateway ||
		old.PeerExternalGatewayInterface != new.PeerExternalGatewayInterface ||
		old.PeerGcpGateway != new.PeerGcpGateway ||
		old.Router != new.Router ||
		old.SharedSecretHash != new.SharedSecretHash ||
		old.VpnGateway != new.VpnGateway ||
		old.TargetVpnGateway != new.TargetVpnGateway ||
		old.VpnGatewayInterface != new.VpnGatewayInterface ||
		jsonb.Changed(old.LocalTrafficSelectorJSON, new.LocalTrafficSelectorJSON) ||
		jsonb.Changed(old.RemoteTrafficSelectorJSON, new.RemoteTrafficSelectorJSON)
}

func diffLabels(old, new []bronze.GCPComputeVpnTunnelLabel) ChildDiff {
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
