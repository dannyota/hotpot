package logexclusion

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// ExclusionDiff represents changes between old and new log exclusion states.
type ExclusionDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffExclusionData compares old Ent entity and new data.
func DiffExclusionData(old *ent.BronzeGCPLoggingLogExclusion, new *LogExclusionData) *ExclusionDiff {
	if old == nil {
		return &ExclusionDiff{IsNew: true}
	}
	return &ExclusionDiff{
		IsChanged: hasFieldsChanged(old, new),
	}
}

// HasAnyChange returns true if any part of the log exclusion changed.
func (d *ExclusionDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

func hasFieldsChanged(old *ent.BronzeGCPLoggingLogExclusion, new *LogExclusionData) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Filter != new.Filter ||
		old.Disabled != new.Disabled ||
		old.CreateTime != new.CreateTime ||
		old.UpdateTime != new.UpdateTime
}
