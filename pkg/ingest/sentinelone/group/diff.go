package group

import "github.com/dannyota/hotpot/pkg/storage/ent"

// GroupDiff represents changes between old and new group states.
type GroupDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffGroupData compares old Ent entity and new data.
func DiffGroupData(old *ent.BronzeS1Group, new *GroupData) *GroupDiff {
	if old == nil {
		return &GroupDiff{IsNew: true}
	}

	changed := old.Name != new.Name ||
		old.SiteID != new.SiteID ||
		old.Type != new.Type ||
		old.IsDefault != new.IsDefault ||
		old.Inherits != new.Inherits ||
		old.TotalAgents != new.TotalAgents ||
		old.Creator != new.Creator ||
		old.CreatorID != new.CreatorID ||
		old.FilterName != new.FilterName ||
		old.FilterID != new.FilterID ||
		!nillableIntEqual(old.Rank, new.Rank)

	return &GroupDiff{IsChanged: changed}
}

func nillableIntEqual(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
