package logbucket

import (
	"bytes"

	entlogging "github.com/dannyota/hotpot/pkg/storage/ent/gcp/logging"
)

// LogBucketDiff represents changes between old and new log bucket states.
type LogBucketDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffLogBucketData compares old Ent entity and new data.
func DiffLogBucketData(old *entlogging.BronzeGCPLoggingBucket, new *LogBucketData) *LogBucketDiff {
	if old == nil {
		return &LogBucketDiff{IsNew: true}
	}
	return &LogBucketDiff{
		IsChanged: hasFieldsChanged(old, new),
	}
}

// HasAnyChange returns true if any part of the log bucket changed.
func (d *LogBucketDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

func hasFieldsChanged(old *entlogging.BronzeGCPLoggingBucket, new *LogBucketData) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.RetentionDays != new.RetentionDays ||
		old.Locked != new.Locked ||
		old.LifecycleState != new.LifecycleState ||
		old.AnalyticsEnabled != new.AnalyticsEnabled ||
		old.Location != new.Location ||
		!bytes.Equal(old.CmekSettingsJSON, new.CmekSettingsJSON) ||
		!bytes.Equal(old.IndexConfigsJSON, new.IndexConfigsJSON)
}
