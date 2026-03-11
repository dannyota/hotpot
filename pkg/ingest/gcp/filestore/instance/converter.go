package instance

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
)

// InstanceData holds converted Filestore instance data ready for Ent insertion.
type InstanceData struct {
	ID                   string
	Name                 string
	Description          string
	State                int
	StatusMessage        string
	CreateTime           string
	Tier                 int
	LabelsJSON           json.RawMessage
	FileSharesJSON       json.RawMessage
	NetworksJSON         json.RawMessage
	Etag                 string
	SatisfiesPzs         bool
	SatisfiesPzi         bool
	KmsKeyName           string
	SuspensionReasonsJSON json.RawMessage
	MaxCapacityGB        int64
	Protocol             int
	ProjectID            string
	Location             string
	CollectedAt          time.Time
}

// ConvertInstance converts a GCP API Filestore Instance to Ent-compatible data.
func ConvertInstance(inst *filestorepb.Instance, projectID string, collectedAt time.Time) (*InstanceData, error) {
	data := &InstanceData{
		ID:            inst.GetName(),
		Name:          inst.GetName(),
		Description:   inst.GetDescription(),
		State:         int(inst.GetState()),
		StatusMessage: inst.GetStatusMessage(),
		Tier:          int(inst.GetTier()),
		Etag:          inst.GetEtag(),
		SatisfiesPzs:  inst.GetSatisfiesPzs().GetValue(),
		SatisfiesPzi:  inst.GetSatisfiesPzi(),
		KmsKeyName:    inst.GetKmsKeyName(),
		Protocol:      int(inst.GetProtocol()),
		ProjectID:     projectID,
		Location:      extractLocation(inst.GetName()),
		CollectedAt:   collectedAt,
	}

	if inst.GetCreateTime() != nil {
		data.CreateTime = inst.GetCreateTime().AsTime().Format(time.RFC3339)
	}

	if len(inst.GetLabels()) > 0 {
		j, err := json.Marshal(inst.GetLabels())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal labels for instance %s: %w", inst.GetName(), err)
		}
		data.LabelsJSON = j
	}

	if len(inst.GetFileShares()) > 0 {
		j, err := json.Marshal(inst.GetFileShares())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal file_shares for instance %s: %w", inst.GetName(), err)
		}
		data.FileSharesJSON = j
	}

	if len(inst.GetNetworks()) > 0 {
		j, err := json.Marshal(inst.GetNetworks())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal networks for instance %s: %w", inst.GetName(), err)
		}
		data.NetworksJSON = j
	}

	if len(inst.GetSuspensionReasons()) > 0 {
		j, err := json.Marshal(inst.GetSuspensionReasons())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal suspension_reasons for instance %s: %w", inst.GetName(), err)
		}
		data.SuspensionReasonsJSON = j
	}

	return data, nil
}

// extractLocation extracts the location from a resource name.
// Format: projects/{project}/locations/{location}/instances/{instance}
func extractLocation(resourceName string) string {
	parts := strings.Split(resourceName, "/")
	if len(parts) >= 4 {
		return parts[3]
	}
	return ""
}
