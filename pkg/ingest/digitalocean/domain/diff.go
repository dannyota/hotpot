package domain

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// DomainDiff represents changes between old and new Domain states.
type DomainDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffDomainData compares old Ent entity and new data.
func DiffDomainData(old *ent.BronzeDODomain, new *DomainData) *DomainDiff {
	if old == nil {
		return &DomainDiff{IsNew: true}
	}

	changed := old.TTL != new.TTL ||
		old.ZoneFile != new.ZoneFile

	return &DomainDiff{IsChanged: changed}
}

// DomainRecordDiff represents changes between old and new Domain Record states.
type DomainRecordDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffDomainRecordData compares old Ent entity and new data.
func DiffDomainRecordData(old *ent.BronzeDODomainRecord, new *DomainRecordData) *DomainRecordDiff {
	if old == nil {
		return &DomainRecordDiff{IsNew: true}
	}

	changed := old.DomainName != new.DomainName ||
		old.RecordID != new.RecordID ||
		old.Type != new.Type ||
		old.Name != new.Name ||
		old.Data != new.Data ||
		old.Priority != new.Priority ||
		old.Port != new.Port ||
		old.TTL != new.TTL ||
		old.Weight != new.Weight ||
		old.Flags != new.Flags ||
		old.Tag != new.Tag

	return &DomainRecordDiff{IsChanged: changed}
}
