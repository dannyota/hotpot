package address

import (
	"bytes"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// AddressDiff represents changes between old and new address states.
type AddressDiff struct {
	IsNew     bool
	IsChanged bool

	// Child diffs (for granular tracking)
	LabelsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffAddressData compares old Ent entity and new AddressData.
func DiffAddressData(old *ent.BronzeGCPComputeAddress, new *AddressData) *AddressDiff {
	if old == nil {
		return &AddressDiff{
			IsNew:      true,
			LabelsDiff: ChildDiff{Changed: true},
		}
	}

	diff := &AddressDiff{}

	// Compare address-level fields
	diff.IsChanged = hasAddressFieldsChanged(old, new)

	// Compare children (note: old.Edges.Labels may be nil if not loaded)
	oldLabels := old.Edges.Labels
	diff.LabelsDiff = diffLabels(oldLabels, new.Labels)

	return diff
}

// HasAnyChange returns true if any part of the address changed.
func (d *AddressDiff) HasAnyChange() bool {
	if d.IsNew || d.IsChanged {
		return true
	}
	return d.LabelsDiff.Changed
}

// hasAddressFieldsChanged compares address-level fields (excluding children).
func hasAddressFieldsChanged(old *ent.BronzeGCPComputeAddress, new *AddressData) bool {
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

func diffLabels(old []*ent.BronzeGCPComputeAddressLabel, new []AddressLabelData) ChildDiff {
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
