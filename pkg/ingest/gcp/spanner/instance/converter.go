package instance

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
)

// InstanceData holds converted Spanner instance data ready for Ent insertion.
type InstanceData struct {
	ResourceID                string
	Name                      string
	Config                    string
	DisplayName               string
	NodeCount                 int32
	ProcessingUnits           int32
	State                     int
	LabelsJSON                json.RawMessage
	EndpointUrisJSON          json.RawMessage
	CreateTime                string
	UpdateTime                string
	Edition                   int
	DefaultBackupScheduleType int
	ProjectID                 string
	CollectedAt               time.Time
}

// ConvertInstance converts a GCP API Spanner instance to InstanceData.
func ConvertInstance(inst *instancepb.Instance, projectID string, collectedAt time.Time) (*InstanceData, error) {
	data := &InstanceData{
		ResourceID:                inst.GetName(),
		Name:                      inst.GetName(),
		Config:                    inst.GetConfig(),
		DisplayName:               inst.GetDisplayName(),
		NodeCount:                 inst.GetNodeCount(),
		ProcessingUnits:           inst.GetProcessingUnits(),
		State:                     int(inst.GetState()),
		Edition:                   int(inst.GetEdition()),
		DefaultBackupScheduleType: int(inst.GetDefaultBackupScheduleType()),
		ProjectID:                 projectID,
		CollectedAt:               collectedAt,
	}

	// Convert create_time
	if inst.GetCreateTime() != nil {
		data.CreateTime = inst.GetCreateTime().AsTime().Format(time.RFC3339)
	}

	// Convert update_time
	if inst.GetUpdateTime() != nil {
		data.UpdateTime = inst.GetUpdateTime().AsTime().Format(time.RFC3339)
	}

	// Convert labels to JSON
	if len(inst.GetLabels()) > 0 {
		labelsBytes, err := json.Marshal(inst.GetLabels())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal labels for instance %s: %w", inst.GetName(), err)
		}
		data.LabelsJSON = labelsBytes
	}

	// Convert endpoint URIs to JSON
	if len(inst.GetEndpointUris()) > 0 {
		endpointBytes, err := json.Marshal(inst.GetEndpointUris())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal endpoint_uris for instance %s: %w", inst.GetName(), err)
		}
		data.EndpointUrisJSON = endpointBytes
	}

	return data, nil
}
