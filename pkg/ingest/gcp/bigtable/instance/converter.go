package instance

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/bigtable"
)

// InstanceData holds converted Bigtable instance data ready for Ent insertion.
type InstanceData struct {
	ID           string
	DisplayName  string
	State        int32
	InstanceType int32
	CreateTime   string
	SatisfiesPzs *bool
	LabelsJSON   json.RawMessage
	ProjectID    string
	CollectedAt  time.Time
}

// instanceStateToInt32 converts bigtable instance state string to int32.
func instanceStateToInt32(info *bigtable.InstanceInfo) int32 {
	// InstanceInfo doesn't expose raw proto state; map from the typed field.
	// The bigtable Go client doesn't expose numeric state directly, so we use
	// the InstanceState type which is an int underlying type.
	return int32(info.InstanceState)
}

// instanceTypeToInt32 converts bigtable instance type to int32.
func instanceTypeToInt32(info *bigtable.InstanceInfo) int32 {
	return int32(info.InstanceType)
}

// ConvertInstance converts a raw GCP Bigtable InstanceInfo to Ent-compatible data.
func ConvertInstance(info *bigtable.InstanceInfo, projectID string, collectedAt time.Time) (*InstanceData, error) {
	data := &InstanceData{
		ID:           fmt.Sprintf("projects/%s/instances/%s", projectID, info.Name),
		DisplayName:  info.DisplayName,
		State:        instanceStateToInt32(info),
		InstanceType: instanceTypeToInt32(info),
		ProjectID:    projectID,
		CollectedAt:  collectedAt,
	}

	if len(info.Labels) > 0 {
		labelsJSON, err := json.Marshal(info.Labels)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal labels: %w", err)
		}
		data.LabelsJSON = labelsJSON
	}

	return data, nil
}
