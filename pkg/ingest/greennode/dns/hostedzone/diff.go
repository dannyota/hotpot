package hostedzone

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// HostedZoneDiff represents changes between old and new hosted zone states.
type HostedZoneDiff struct {
	IsNew     bool
	IsChanged bool

	RecordsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffHostedZoneData compares old Ent entity and new HostedZoneData.
func DiffHostedZoneData(old *ent.BronzeGreenNodeDNSHostedZone, new *HostedZoneData) *HostedZoneDiff {
	if old == nil {
		return &HostedZoneDiff{
			IsNew:       true,
			RecordsDiff: ChildDiff{Changed: true},
		}
	}

	diff := &HostedZoneDiff{}
	diff.IsChanged = hasHostedZoneFieldsChanged(old, new)
	diff.RecordsDiff = diffRecords(old.Edges.Records, new.Records)

	return diff
}

// HasAnyChange returns true if any part of the hosted zone changed.
func (d *HostedZoneDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.RecordsDiff.Changed
}

func hasHostedZoneFieldsChanged(old *ent.BronzeGreenNodeDNSHostedZone, new *HostedZoneData) bool {
	return old.DomainName != new.DomainName ||
		old.Status != new.Status ||
		old.Description != new.Description ||
		old.Type != new.Type ||
		old.CountRecords != new.CountRecords ||
		!bytes.Equal(old.AssocVpcIdsJSON, new.AssocVpcIdsJSON) ||
		!bytes.Equal(old.AssocVpcMapRegionJSON, new.AssocVpcMapRegionJSON) ||
		old.PortalUserID != new.PortalUserID ||
		old.CreatedAtAPI != new.CreatedAtAPI ||
		!nillableStringEqual(old.DeletedAtAPI, new.DeletedAtAPI) ||
		old.UpdatedAtAPI != new.UpdatedAtAPI
}

func nillableStringEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func diffRecords(old []*ent.BronzeGreenNodeDNSRecord, new []RecordData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	oldMap := make(map[string]*ent.BronzeGreenNodeDNSRecord)
	for _, r := range old {
		oldMap[r.RecordID] = r
	}
	for _, r := range new {
		oldRec, ok := oldMap[r.RecordID]
		if !ok {
			return ChildDiff{Changed: true}
		}
		if hasRecordFieldsChanged(oldRec, &r) {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}

func hasRecordFieldsChanged(old *ent.BronzeGreenNodeDNSRecord, new *RecordData) bool {
	return old.SubDomain != new.SubDomain ||
		old.Status != new.Status ||
		old.Type != new.Type ||
		old.RoutingPolicy != new.RoutingPolicy ||
		!bytes.Equal(old.ValueJSON, new.ValueJSON) ||
		old.TTL != new.TTL ||
		!nillableBoolEqual(old.EnableStickySession, new.EnableStickySession) ||
		old.CreatedAtAPI != new.CreatedAtAPI ||
		!nillableStringEqual(old.DeletedAtAPI, new.DeletedAtAPI) ||
		old.UpdatedAtAPI != new.UpdatedAtAPI
}

func nillableBoolEqual(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
