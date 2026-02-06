package address

import (
	"hotpot/pkg/base/jsonb"
	"hotpot/pkg/base/models/bronze"
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

// DiffAddress compares old and new address states.
// Returns nil if old is nil (new address).
func DiffAddress(old, new *bronze.GCPComputeAddress) *AddressDiff {
	if old == nil {
		return &AddressDiff{
			IsNew:      true,
			LabelsDiff: ChildDiff{Changed: true},
		}
	}

	diff := &AddressDiff{}

	// Compare address-level fields
	diff.IsChanged = hasAddressFieldsChanged(old, new)

	// Compare children
	diff.LabelsDiff = diffLabels(old.Labels, new.Labels)

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
func hasAddressFieldsChanged(old, new *bronze.GCPComputeAddress) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Address != new.Address ||
		old.AddressType != new.AddressType ||
		old.IpVersion != new.IpVersion ||
		old.Ipv6EndpointType != new.Ipv6EndpointType ||
		old.IpCollection != new.IpCollection ||
		old.Region != new.Region ||
		old.Status != new.Status ||
		old.Purpose != new.Purpose ||
		old.Network != new.Network ||
		old.Subnetwork != new.Subnetwork ||
		old.NetworkTier != new.NetworkTier ||
		old.PrefixLength != new.PrefixLength ||
		old.LabelFingerprint != new.LabelFingerprint ||
		jsonb.Changed(old.UsersJSON, new.UsersJSON)
}

func diffLabels(old, new []bronze.GCPComputeAddressLabel) ChildDiff {
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
