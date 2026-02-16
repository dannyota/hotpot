package function

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/functions/apiv2/functionspb"
)

// FunctionData holds converted Cloud Function data ready for Ent insertion.
type FunctionData struct {
	ID               string
	Name             string
	Description      string
	Environment      int
	State            int
	BuildConfigJSON  json.RawMessage
	ServiceConfigJSON json.RawMessage
	EventTriggerJSON json.RawMessage
	StateMessagesJSON json.RawMessage
	UpdateTime       string
	CreateTime       string
	LabelsJSON       json.RawMessage
	KmsKeyName       string
	URL              string
	SatisfiesPzs     bool
	ProjectID        string
	Location         string
	CollectedAt      time.Time
}

// ConvertFunction converts a GCP API Function to Ent-compatible data.
func ConvertFunction(f *functionspb.Function, projectID string, collectedAt time.Time) (*FunctionData, error) {
	data := &FunctionData{
		ID:           f.GetName(),
		Name:         f.GetName(),
		Description:  f.GetDescription(),
		Environment:  int(f.GetEnvironment()),
		State:        int(f.GetState()),
		KmsKeyName:   f.GetKmsKeyName(),
		URL:          f.GetUrl(),
		SatisfiesPzs: f.GetSatisfiesPzs(),
		ProjectID:    projectID,
		Location:     extractLocation(f.GetName()),
		CollectedAt:  collectedAt,
	}

	if f.GetUpdateTime() != nil {
		data.UpdateTime = f.GetUpdateTime().AsTime().Format(time.RFC3339)
	}
	if f.GetCreateTime() != nil {
		data.CreateTime = f.GetCreateTime().AsTime().Format(time.RFC3339)
	}

	if f.GetBuildConfig() != nil {
		j, err := json.Marshal(f.GetBuildConfig())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal build_config for function %s: %w", f.GetName(), err)
		}
		data.BuildConfigJSON = j
	}

	if f.GetServiceConfig() != nil {
		j, err := json.Marshal(f.GetServiceConfig())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal service_config for function %s: %w", f.GetName(), err)
		}
		data.ServiceConfigJSON = j
	}

	if f.GetEventTrigger() != nil {
		j, err := json.Marshal(f.GetEventTrigger())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal event_trigger for function %s: %w", f.GetName(), err)
		}
		data.EventTriggerJSON = j
	}

	if len(f.GetStateMessages()) > 0 {
		j, err := json.Marshal(f.GetStateMessages())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal state_messages for function %s: %w", f.GetName(), err)
		}
		data.StateMessagesJSON = j
	}

	if len(f.GetLabels()) > 0 {
		j, err := json.Marshal(f.GetLabels())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal labels for function %s: %w", f.GetName(), err)
		}
		data.LabelsJSON = j
	}

	return data, nil
}

// extractLocation extracts the location from a function resource name.
// Format: projects/{project}/locations/{location}/functions/{function}
func extractLocation(name string) string {
	parts := strings.Split(name, "/")
	for i, p := range parts {
		if p == "locations" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
