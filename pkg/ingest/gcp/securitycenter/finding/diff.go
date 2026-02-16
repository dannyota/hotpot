package finding

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// FindingDiff represents changes between old and new SCC finding state.
type FindingDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *FindingDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffFindingData compares existing Ent entity with new FindingData and returns differences.
func DiffFindingData(old *ent.BronzeGCPSecurityCenterFinding, new *FindingData) *FindingDiff {
	diff := &FindingDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.State != new.State ||
		old.Category != new.Category ||
		old.Severity != new.Severity ||
		old.FindingClass != new.FindingClass ||
		old.Mute != new.Mute ||
		old.EventTime != new.EventTime ||
		old.ExternalURI != new.ExternalURI ||
		old.ResourceName != new.ResourceName ||
		old.CanonicalName != new.CanonicalName {
		diff.IsChanged = true
	}

	return diff
}
