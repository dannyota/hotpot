package targetvpngateway

import (
	"bytes"

	"hotpot/pkg/storage/ent"
)

// TargetVpnGatewayDiff represents changes between old and new target VPN gateway states.
type TargetVpnGatewayDiff struct {
	IsNew      bool
	IsChanged  bool
	LabelsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	HasChanges bool
}

// HasAnyChange returns true if any part of the target VPN gateway changed.
func (d *TargetVpnGatewayDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.LabelsDiff.HasChanges
}

// DiffTargetVpnGatewayData compares existing Ent entity with new TargetVpnGatewayData.
func DiffTargetVpnGatewayData(old *ent.BronzeGCPVPNTargetGateway, new *TargetVpnGatewayData) *TargetVpnGatewayDiff {
	diff := &TargetVpnGatewayDiff{}

	// New target VPN gateway
	if old == nil {
		diff.IsNew = true
		return diff
	}

	// Compare core fields
	diff.IsChanged = hasTargetVpnGatewayFieldsChanged(old, new)

	// Compare labels
	diff.LabelsDiff = diffLabelsData(old.Edges.Labels, new.Labels)

	return diff
}

// hasTargetVpnGatewayFieldsChanged compares target VPN gateway-level fields (excluding children).
func hasTargetVpnGatewayFieldsChanged(old *ent.BronzeGCPVPNTargetGateway, new *TargetVpnGatewayData) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Status != new.Status ||
		old.Region != new.Region ||
		old.Network != new.Network ||
		old.SelfLink != new.SelfLink ||
		old.CreationTimestamp != new.CreationTimestamp ||
		old.LabelFingerprint != new.LabelFingerprint ||
		!bytes.Equal(old.ForwardingRulesJSON, new.ForwardingRulesJSON) ||
		!bytes.Equal(old.TunnelsJSON, new.TunnelsJSON)
}

func diffLabelsData(old []*ent.BronzeGCPVPNTargetGatewayLabel, new []LabelData) ChildDiff {
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
