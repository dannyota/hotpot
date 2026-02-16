package managedzone

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// ManagedZoneDiff represents changes between old and new managed zone states.
type ManagedZoneDiff struct {
	IsNew     bool
	IsChanged bool

	// Child diffs (for granular tracking)
	LabelDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffManagedZoneData compares old Ent entity and new data.
func DiffManagedZoneData(old *ent.BronzeGCPDNSManagedZone, new *ManagedZoneData) *ManagedZoneDiff {
	if old == nil {
		return &ManagedZoneDiff{
			IsNew:     true,
			LabelDiff: ChildDiff{Changed: true},
		}
	}

	diff := &ManagedZoneDiff{}

	// Compare managed zone-level fields
	diff.IsChanged = hasManagedZoneFieldsChanged(old, new)

	// Compare label children
	var oldLabels []*ent.BronzeGCPDNSManagedZoneLabel
	if old.Edges.Labels != nil {
		oldLabels = old.Edges.Labels
	}
	diff.LabelDiff = diffLabels(oldLabels, new.Labels)

	return diff
}

// HasAnyChange returns true if any part of the managed zone changed.
func (d *ManagedZoneDiff) HasAnyChange() bool {
	if d.IsNew || d.IsChanged {
		return true
	}
	return d.LabelDiff.Changed
}

// hasManagedZoneFieldsChanged compares managed zone-level fields (excluding children).
func hasManagedZoneFieldsChanged(old *ent.BronzeGCPDNSManagedZone, new *ManagedZoneData) bool {
	return old.Name != new.Name ||
		old.DNSName != new.DnsName ||
		old.Description != new.Description ||
		old.Visibility != new.Visibility ||
		old.CreationTime != new.CreationTime ||
		!bytes.Equal(old.DnssecConfigJSON, new.DnssecConfigJSON) ||
		!bytes.Equal(old.PrivateVisibilityConfigJSON, new.PrivateVisibilityConfigJSON) ||
		!bytes.Equal(old.ForwardingConfigJSON, new.ForwardingConfigJSON) ||
		!bytes.Equal(old.PeeringConfigJSON, new.PeeringConfigJSON) ||
		!bytes.Equal(old.CloudLoggingConfigJSON, new.CloudLoggingConfigJSON)
}

func diffLabels(old []*ent.BronzeGCPDNSManagedZoneLabel, new []LabelData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}

	// Build map of old labels by key
	oldMap := make(map[string]string)
	for _, l := range old {
		oldMap[l.Key] = l.Value
	}

	// Compare each new label
	for _, newL := range new {
		oldVal, ok := oldMap[newL.Key]
		if !ok || oldVal != newL.Value {
			return ChildDiff{Changed: true}
		}
	}

	return ChildDiff{Changed: false}
}
