package vpngateway

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// VpnGatewayDiff represents changes between old and new VPN gateway states.
type VpnGatewayDiff struct {
	IsNew      bool
	IsChanged  bool
	LabelsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	HasChanges bool
}

// HasAnyChange returns true if any part of the VPN gateway changed.
func (d *VpnGatewayDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.LabelsDiff.HasChanges
}

// DiffVpnGatewayData compares existing Ent entity with new VpnGatewayData.
func DiffVpnGatewayData(old *ent.BronzeGCPVPNGateway, new *VpnGatewayData) *VpnGatewayDiff {
	diff := &VpnGatewayDiff{}

	// New VPN gateway
	if old == nil {
		diff.IsNew = true
		return diff
	}

	// Compare core fields
	diff.IsChanged = hasVpnGatewayFieldsChanged(old, new)

	// Compare labels
	diff.LabelsDiff = diffLabelsData(old.Edges.Labels, new.Labels)

	return diff
}

// hasVpnGatewayFieldsChanged compares VPN gateway-level fields (excluding children).
func hasVpnGatewayFieldsChanged(old *ent.BronzeGCPVPNGateway, new *VpnGatewayData) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Region != new.Region ||
		old.Network != new.Network ||
		old.SelfLink != new.SelfLink ||
		old.CreationTimestamp != new.CreationTimestamp ||
		old.LabelFingerprint != new.LabelFingerprint ||
		old.GatewayIPVersion != new.GatewayIpVersion ||
		old.StackType != new.StackType ||
		!bytes.Equal(old.VpnInterfacesJSON, new.VpnInterfacesJSON)
}

func diffLabelsData(old []*ent.BronzeGCPVPNGatewayLabel, new []LabelData) ChildDiff {
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
