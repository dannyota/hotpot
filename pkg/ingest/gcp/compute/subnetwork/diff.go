package subnetwork

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// SubnetworkDiff represents changes between old and new subnetwork states.
type SubnetworkDiff struct {
	IsNew     bool
	IsChanged bool

	// Child diffs (for granular tracking)
	SecondaryRangesDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffSubnetworkData compares old Ent entity and new data.
func DiffSubnetworkData(old *ent.BronzeGCPComputeSubnetwork, new *SubnetworkData) *SubnetworkDiff {
	if old == nil {
		return &SubnetworkDiff{
			IsNew:               true,
			SecondaryRangesDiff: ChildDiff{Changed: true},
		}
	}

	diff := &SubnetworkDiff{}

	// Compare subnetwork-level fields
	diff.IsChanged = hasSubnetworkFieldsChanged(old, new)

	// Compare children
	var oldRanges []*ent.BronzeGCPComputeSubnetworkSecondaryRange
	if old.Edges.SecondaryIPRanges != nil {
		oldRanges = old.Edges.SecondaryIPRanges
	}
	diff.SecondaryRangesDiff = diffSecondaryRanges(oldRanges, new.SecondaryIpRanges)

	return diff
}

// HasAnyChange returns true if any part of the subnetwork changed.
func (d *SubnetworkDiff) HasAnyChange() bool {
	if d.IsNew || d.IsChanged {
		return true
	}
	return d.SecondaryRangesDiff.Changed
}

// hasSubnetworkFieldsChanged compares subnetwork-level fields (excluding children).
func hasSubnetworkFieldsChanged(old *ent.BronzeGCPComputeSubnetwork, new *SubnetworkData) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Network != new.Network ||
		old.Region != new.Region ||
		old.IPCidrRange != new.IpCidrRange ||
		old.GatewayAddress != new.GatewayAddress ||
		old.Purpose != new.Purpose ||
		old.Role != new.Role ||
		old.PrivateIPGoogleAccess != new.PrivateIpGoogleAccess ||
		old.PrivateIpv6GoogleAccess != new.PrivateIpv6GoogleAccess ||
		old.StackType != new.StackType ||
		old.Ipv6AccessType != new.Ipv6AccessType ||
		old.InternalIpv6Prefix != new.InternalIpv6Prefix ||
		old.ExternalIpv6Prefix != new.ExternalIpv6Prefix ||
		!bytes.Equal(old.LogConfigJSON, new.LogConfigJSON) ||
		old.Fingerprint != new.Fingerprint
}

func diffSecondaryRanges(old []*ent.BronzeGCPComputeSubnetworkSecondaryRange, new []SecondaryRangeData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}

	// Build map of old ranges by name
	oldMap := make(map[string]string)
	for _, r := range old {
		oldMap[r.RangeName] = r.IPCidrRange
	}

	// Compare each new range
	for _, newR := range new {
		oldCidr, ok := oldMap[newR.RangeName]
		if !ok || oldCidr != newR.IpCidrRange {
			return ChildDiff{Changed: true}
		}
	}

	return ChildDiff{Changed: false}
}
