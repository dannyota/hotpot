package vpntunnel

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// VpnTunnelDiff represents changes between old and new VPN tunnel states.
type VpnTunnelDiff struct {
	IsNew      bool
	IsChanged  bool
	LabelsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	HasChanges bool
}

// HasAnyChange returns true if any part of the VPN tunnel changed.
func (d *VpnTunnelDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.LabelsDiff.HasChanges
}

// DiffVpnTunnelData compares existing Ent entity with new VpnTunnelData.
func DiffVpnTunnelData(old *ent.BronzeGCPVPNTunnel, new *VpnTunnelData) *VpnTunnelDiff {
	diff := &VpnTunnelDiff{}

	// New VPN tunnel
	if old == nil {
		diff.IsNew = true
		return diff
	}

	// Compare core fields
	diff.IsChanged = hasVpnTunnelFieldsChanged(old, new)

	// Compare labels
	diff.LabelsDiff = diffLabelsData(old.Edges.Labels, new.Labels)

	return diff
}

// hasVpnTunnelFieldsChanged compares VPN tunnel-level fields (excluding children).
func hasVpnTunnelFieldsChanged(old *ent.BronzeGCPVPNTunnel, new *VpnTunnelData) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Status != new.Status ||
		old.DetailedStatus != new.DetailedStatus ||
		old.Region != new.Region ||
		old.SelfLink != new.SelfLink ||
		old.CreationTimestamp != new.CreationTimestamp ||
		old.LabelFingerprint != new.LabelFingerprint ||
		old.IkeVersion != new.IkeVersion ||
		old.PeerIP != new.PeerIp ||
		old.PeerExternalGateway != new.PeerExternalGateway ||
		old.PeerExternalGatewayInterface != new.PeerExternalGatewayInterface ||
		old.PeerGcpGateway != new.PeerGcpGateway ||
		old.Router != new.Router ||
		old.SharedSecretHash != new.SharedSecretHash ||
		old.VpnGateway != new.VpnGateway ||
		old.TargetVpnGateway != new.TargetVpnGateway ||
		old.VpnGatewayInterface != new.VpnGatewayInterface ||
		!bytes.Equal(old.LocalTrafficSelectorJSON, new.LocalTrafficSelectorJSON) ||
		!bytes.Equal(old.RemoteTrafficSelectorJSON, new.RemoteTrafficSelectorJSON)
}

func diffLabelsData(old []*ent.BronzeGCPVPNTunnelLabel, new []LabelData) ChildDiff {
	diff := ChildDiff{}

	if len(old) != len(new) {
		diff.HasChanges = true
		return diff
	}

	// Build map of old labels
	oldMap := make(map[string]string, len(old))
	for _, l := range old {
		oldMap[l.Key] = l.Value
	}

	// Compare with new labels
	for _, l := range new {
		if oldValue, ok := oldMap[l.Key]; !ok || oldValue != l.Value {
			diff.HasChanges = true
			return diff
		}
	}

	return diff
}
