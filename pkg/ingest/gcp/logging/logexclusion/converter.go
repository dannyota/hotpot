package logexclusion

import (
	"time"

	"cloud.google.com/go/logging/apiv2/loggingpb"
)

// LogExclusionData holds converted log exclusion data ready for Ent insertion.
type LogExclusionData struct {
	ResourceID  string
	Name        string
	Description string
	Filter      string
	Disabled    bool
	CreateTime  string
	UpdateTime  string
	ProjectID   string
	CollectedAt time.Time
}

// ConvertExclusion converts a GCP API LogExclusion to Ent-compatible data.
func ConvertExclusion(e *loggingpb.LogExclusion, projectID string, collectedAt time.Time) *LogExclusionData {
	data := &LogExclusionData{
		ResourceID:  e.GetName(),
		Name:        e.GetName(),
		Description: e.GetDescription(),
		Filter:      e.GetFilter(),
		Disabled:    e.GetDisabled(),
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}

	if e.GetCreateTime() != nil {
		data.CreateTime = e.GetCreateTime().AsTime().Format(time.RFC3339)
	}
	if e.GetUpdateTime() != nil {
		data.UpdateTime = e.GetUpdateTime().AsTime().Format(time.RFC3339)
	}

	return data
}
