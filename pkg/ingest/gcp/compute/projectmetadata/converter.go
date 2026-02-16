package projectmetadata

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// ProjectMetadataData holds converted project metadata data ready for Ent insertion.
type ProjectMetadataData struct {
	ID                      string
	Name                    string
	DefaultServiceAccount   string
	DefaultNetworkTier      string
	XpnProjectStatus        string
	CreationTimestamp       string
	UsageExportLocationJSON json.RawMessage
	ProjectID               string
	CollectedAt             time.Time
	Items                   []ItemData
}

// ItemData holds converted metadata item data.
type ItemData struct {
	Key   string
	Value string
}

// ConvertProjectMetadata converts a GCP API Project to Ent-compatible data.
// Preserves raw API data with minimal transformation.
func ConvertProjectMetadata(project *computepb.Project, projectID string, collectedAt time.Time) (*ProjectMetadataData, error) {
	data := &ProjectMetadataData{
		ID:                    fmt.Sprintf("%d", project.GetId()),
		Name:                  project.GetName(),
		DefaultServiceAccount: project.GetDefaultServiceAccount(),
		DefaultNetworkTier:    project.GetDefaultNetworkTier(),
		XpnProjectStatus:     project.GetXpnProjectStatus(),
		CreationTimestamp:     project.GetCreationTimestamp(),
		ProjectID:             projectID,
		CollectedAt:           collectedAt,
	}

	// Marshal usage export location to JSON
	if project.GetUsageExportLocation() != nil {
		usageJSON, err := json.Marshal(project.GetUsageExportLocation())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal usage_export_location for project %s: %w", project.GetName(), err)
		}
		data.UsageExportLocationJSON = usageJSON
	}

	// Convert common instance metadata items
	data.Items = ConvertItems(project.GetCommonInstanceMetadata().GetItems())

	return data, nil
}

// ConvertItems converts metadata items from GCP API to data structs.
func ConvertItems(items []*computepb.Items) []ItemData {
	if len(items) == 0 {
		return nil
	}

	result := make([]ItemData, 0, len(items))
	for _, item := range items {
		result = append(result, ItemData{
			Key:   item.GetKey(),
			Value: item.GetValue(),
		})
	}

	return result
}
