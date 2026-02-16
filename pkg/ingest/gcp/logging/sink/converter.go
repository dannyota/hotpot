package sink

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/logging/apiv2/loggingpb"
)

// SinkData holds converted sink data ready for Ent insertion.
type SinkData struct {
	ResourceID        string
	Name              string
	Destination       string
	Filter            string
	Description       string
	Disabled          bool
	IncludeChildren   bool
	WriterIdentity    string
	ExclusionsJSON    json.RawMessage
	BigqueryOptionsJSON json.RawMessage
	ProjectID         string
	CollectedAt       time.Time
}

// ConvertSink converts a GCP API LogSink to Ent-compatible data.
func ConvertSink(s *loggingpb.LogSink, projectID string, collectedAt time.Time) (*SinkData, error) {
	data := &SinkData{
		ResourceID:      s.GetName(),
		Name:            s.GetName(),
		Destination:     s.GetDestination(),
		Filter:          s.GetFilter(),
		Description:     s.GetDescription(),
		Disabled:        s.GetDisabled(),
		IncludeChildren: s.GetIncludeChildren(),
		WriterIdentity:  s.GetWriterIdentity(),
		ProjectID:       projectID,
		CollectedAt:     collectedAt,
	}

	if s.GetExclusions() != nil {
		exclusionsJSON, err := json.Marshal(s.GetExclusions())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal exclusions for sink %s: %w", s.GetName(), err)
		}
		data.ExclusionsJSON = exclusionsJSON
	}

	if s.GetBigqueryOptions() != nil {
		bqJSON, err := json.Marshal(s.GetBigqueryOptions())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal bigquery_options for sink %s: %w", s.GetName(), err)
		}
		data.BigqueryOptionsJSON = bqJSON
	}

	return data, nil
}
