package enabledservice

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/serviceusage/apiv1/serviceusagepb"
)

// EnabledServiceData holds converted enabled service data ready for Ent insertion.
type EnabledServiceData struct {
	ID          string
	Name        string
	Parent      string
	ConfigJSON  json.RawMessage
	State       int
	ProjectID   string
	CollectedAt time.Time
}

// ConvertEnabledService converts a GCP API Service to Ent-compatible data.
func ConvertEnabledService(s *serviceusagepb.Service, projectID string, collectedAt time.Time) (*EnabledServiceData, error) {
	data := &EnabledServiceData{
		ID:          s.GetName(),
		Name:        s.GetName(),
		Parent:      s.GetParent(),
		State:       int(s.GetState()),
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}

	if s.GetConfig() != nil {
		j, err := json.Marshal(s.GetConfig())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config for service %s: %w", s.GetName(), err)
		}
		data.ConfigJSON = j
	}

	return data, nil
}
