package globaladdress

import (
	"bytes"
	"hotpot/pkg/storage/ent"
)

// GlobalAddressDiff represents changes between old and new global address states.
type GlobalAddressDiff struct {
	IsNew     bool
	IsChanged bool

	// Child diffs (for granular tracking)
	LabelsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffGlobalAddressData compares old Ent entity and new GlobalAddressData.
func DiffGlobalAddressData(old *ent.BronzeGCPComputeGlobalAddress, new *GlobalAddressData) *GlobalAddressDiff {
	if old == nil {
		return &GlobalAddressDiff{
			IsNew:      true,
			LabelsDiff: ChildDiff{Changed: true},
		}
	}

	diff := &GlobalAddressDiff{}

	// Compare address-level fields
	diff.IsChanged = hasGlobalAddressFieldsChanged(old, new)

	// Compare children (note: old.Edges.Labels may be nil if not loaded)
	oldLabels := old.Edges.Labels
	diff.LabelsDiff = diffLabels(oldLabels, new.Labels)

	return diff
}

// HasAnyChange returns true if any part of the global address changed.
func (d *GlobalAddressDiff) HasAnyChange() bool {
	if d.IsNew || d.IsChanged {
		return true
	}
	return d.LabelsDiff.Changed
}

// hasGlobalAddressFieldsChanged compares address-level fields (excluding children).
func hasGlobalAddressFieldsChanged(old *ent.BronzeGCPComputeGlobalAddress, new *GlobalAddressData) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Address != new.Address ||
		old.AddressType != new.AddressType ||
		old.IPVersion != new.IpVersion ||
		old.Ipv6EndpointType != new.Ipv6EndpointType ||
		old.IPCollection != new.IpCollection ||
		old.Region != new.Region ||
		old.Status != new.Status ||
		old.Purpose != new.Purpose ||
		old.Network != new.Network ||
		old.Subnetwork != new.Subnetwork ||
		old.NetworkTier != new.NetworkTier ||
		old.PrefixLength != new.PrefixLength ||
		old.LabelFingerprint != new.LabelFingerprint ||
		!bytes.Equal(old.UsersJSON, new.UsersJSON)
}

func diffLabels(old []*ent.BronzeGCPComputeGlobalAddressLabel, new []GlobalAddressLabelData) ChildDiff {
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
