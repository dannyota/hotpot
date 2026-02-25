package sink

import (
	"bytes"

	entlogging "github.com/dannyota/hotpot/pkg/storage/ent/gcp/logging"
)

// SinkDiff represents changes between old and new sink states.
type SinkDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffSinkData compares old Ent entity and new data.
func DiffSinkData(old *entlogging.BronzeGCPLoggingSink, new *SinkData) *SinkDiff {
	if old == nil {
		return &SinkDiff{IsNew: true}
	}
	return &SinkDiff{
		IsChanged: hasFieldsChanged(old, new),
	}
}

// HasAnyChange returns true if any part of the sink changed.
func (d *SinkDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

func hasFieldsChanged(old *entlogging.BronzeGCPLoggingSink, new *SinkData) bool {
	return old.Name != new.Name ||
		old.Destination != new.Destination ||
		old.Filter != new.Filter ||
		old.Description != new.Description ||
		old.Disabled != new.Disabled ||
		old.IncludeChildren != new.IncludeChildren ||
		old.WriterIdentity != new.WriterIdentity ||
		!bytes.Equal(old.ExclusionsJSON, new.ExclusionsJSON) ||
		!bytes.Equal(old.BigqueryOptionsJSON, new.BigqueryOptionsJSON)
}
