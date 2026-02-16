package firewall

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// FirewallDiff represents changes between old and new firewall states.
type FirewallDiff struct {
	IsNew     bool
	IsChanged bool

	// Child diffs (for granular tracking)
	AllowedDiff ChildDiff
	DeniedDiff  ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffFirewallData compares old Ent entity and new data.
func DiffFirewallData(old *ent.BronzeGCPComputeFirewall, new *FirewallData) *FirewallDiff {
	if old == nil {
		return &FirewallDiff{
			IsNew:       true,
			AllowedDiff: ChildDiff{Changed: true},
			DeniedDiff:  ChildDiff{Changed: true},
		}
	}

	diff := &FirewallDiff{}

	// Compare firewall-level fields
	diff.IsChanged = hasFirewallFieldsChanged(old, new)

	// Compare allowed children (note: old.Edges.Allowed might be nil if not loaded)
	var oldAllowed []*ent.BronzeGCPComputeFirewallAllowed
	if old.Edges.Allowed != nil {
		oldAllowed = old.Edges.Allowed
	}
	diff.AllowedDiff = diffAllowed(oldAllowed, new.Allowed)

	// Compare denied children
	var oldDenied []*ent.BronzeGCPComputeFirewallDenied
	if old.Edges.Denied != nil {
		oldDenied = old.Edges.Denied
	}
	diff.DeniedDiff = diffDenied(oldDenied, new.Denied)

	return diff
}

// HasAnyChange returns true if any part of the firewall changed.
func (d *FirewallDiff) HasAnyChange() bool {
	if d.IsNew || d.IsChanged {
		return true
	}
	return d.AllowedDiff.Changed || d.DeniedDiff.Changed
}

// hasFirewallFieldsChanged compares firewall-level fields (excluding children).
func hasFirewallFieldsChanged(old *ent.BronzeGCPComputeFirewall, new *FirewallData) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Network != new.Network ||
		old.Priority != new.Priority ||
		old.Direction != new.Direction ||
		old.Disabled != new.Disabled ||
		!bytes.Equal(old.SourceRangesJSON, new.SourceRangesJSON) ||
		!bytes.Equal(old.DestinationRangesJSON, new.DestinationRangesJSON) ||
		!bytes.Equal(old.SourceTagsJSON, new.SourceTagsJSON) ||
		!bytes.Equal(old.TargetTagsJSON, new.TargetTagsJSON) ||
		!bytes.Equal(old.SourceServiceAccountsJSON, new.SourceServiceAccountsJSON) ||
		!bytes.Equal(old.TargetServiceAccountsJSON, new.TargetServiceAccountsJSON) ||
		!bytes.Equal(old.LogConfigJSON, new.LogConfigJSON)
}

func diffAllowed(old []*ent.BronzeGCPComputeFirewallAllowed, new []AllowedData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}

	// Build map of old allowed rules by protocol+ports
	type ruleKey struct {
		protocol string
		ports    string
	}
	oldMap := make(map[ruleKey]bool)
	for _, a := range old {
		k := ruleKey{protocol: a.IPProtocol, ports: string(a.PortsJSON)}
		oldMap[k] = true
	}

	// Compare each new allowed rule
	for _, newA := range new {
		k := ruleKey{protocol: newA.IpProtocol, ports: string(newA.PortsJSON)}
		if !oldMap[k] {
			return ChildDiff{Changed: true}
		}
	}

	return ChildDiff{Changed: false}
}

func diffDenied(old []*ent.BronzeGCPComputeFirewallDenied, new []DeniedData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}

	// Build map of old denied rules by protocol+ports
	type ruleKey struct {
		protocol string
		ports    string
	}
	oldMap := make(map[ruleKey]bool)
	for _, d := range old {
		k := ruleKey{protocol: d.IPProtocol, ports: string(d.PortsJSON)}
		oldMap[k] = true
	}

	// Compare each new denied rule
	for _, newD := range new {
		k := ruleKey{protocol: newD.IpProtocol, ports: string(newD.PortsJSON)}
		if !oldMap[k] {
			return ChildDiff{Changed: true}
		}
	}

	return ChildDiff{Changed: false}
}
