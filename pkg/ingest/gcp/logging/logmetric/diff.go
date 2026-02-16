package logmetric

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// LogMetricDiff represents changes between old and new log metric states.
type LogMetricDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffLogMetricData compares old Ent entity and new data.
func DiffLogMetricData(old *ent.BronzeGCPLoggingLogMetric, new *LogMetricData) *LogMetricDiff {
	if old == nil {
		return &LogMetricDiff{IsNew: true}
	}
	return &LogMetricDiff{
		IsChanged: hasFieldsChanged(old, new),
	}
}

// HasAnyChange returns true if any part of the log metric changed.
func (d *LogMetricDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

func hasFieldsChanged(old *ent.BronzeGCPLoggingLogMetric, new *LogMetricData) bool {
	return old.Name != new.Name ||
		old.Description != new.Description ||
		old.Filter != new.Filter ||
		old.ValueExtractor != new.ValueExtractor ||
		old.Version != new.Version ||
		old.Disabled != new.Disabled ||
		old.CreateTime != new.CreateTime ||
		old.UpdateTime != new.UpdateTime ||
		!bytes.Equal(old.MetricDescriptorJSON, new.MetricDescriptorJSON) ||
		!bytes.Equal(old.LabelExtractorsJSON, new.LabelExtractorsJSON) ||
		!bytes.Equal(old.BucketOptionsJSON, new.BucketOptionsJSON)
}
