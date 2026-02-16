package appservice

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/appengine/apiv1/appenginepb"
)

// ServiceData holds converted App Engine service data ready for Ent insertion.
type ServiceData struct {
	ID                  string
	Name                string
	SplitJSON           json.RawMessage
	LabelsJSON          json.RawMessage
	NetworkSettingsJSON json.RawMessage
	ProjectID           string
	CollectedAt         time.Time
}

// ConvertService converts a raw GCP API App Engine service to Ent-compatible data.
func ConvertService(svc *appenginepb.Service, projectID string, collectedAt time.Time) (*ServiceData, error) {
	data := &ServiceData{
		ID:          svc.GetName(),
		Name:        svc.GetName(),
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}

	var err error
	if svc.Split != nil {
		data.SplitJSON, err = json.Marshal(svc.Split)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal split for service %s: %w", svc.GetName(), err)
		}
	}
	if len(svc.Labels) > 0 {
		data.LabelsJSON, err = json.Marshal(svc.Labels)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal labels for service %s: %w", svc.GetName(), err)
		}
	}
	if svc.NetworkSettings != nil {
		data.NetworkSettingsJSON, err = json.Marshal(svc.NetworkSettings)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal network settings for service %s: %w", svc.GetName(), err)
		}
	}

	return data, nil
}
