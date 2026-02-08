package project

import (
	"strings"
	"time"

	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
)

// ProjectData holds converted project data ready for Ent insertion.
type ProjectData struct {
	ID            string
	ProjectNumber string
	DisplayName   string
	State         string
	Parent        string
	Etag          string
	CreateTime    string
	UpdateTime    string
	DeleteTime    string
	Labels        []LabelData
	CollectedAt   time.Time
}

// LabelData holds converted label data.
type LabelData struct {
	Key   string
	Value string
}

// ConvertProject converts a GCP API Project to Ent-compatible data.
// Preserves raw API data with minimal transformation.
func ConvertProject(proj *resourcemanagerpb.Project, collectedAt time.Time) *ProjectData {
	data := &ProjectData{
		ID:            proj.GetProjectId(),
		ProjectNumber: extractProjectNumber(proj.GetName()),
		DisplayName:   proj.GetDisplayName(),
		State:         proj.GetState().String(),
		Parent:        proj.GetParent(),
		Etag:          proj.GetEtag(),
		CollectedAt:   collectedAt,
	}

	// Convert timestamps
	if proj.CreateTime != nil {
		data.CreateTime = proj.CreateTime.AsTime().Format(time.RFC3339)
	}
	if proj.UpdateTime != nil {
		data.UpdateTime = proj.UpdateTime.AsTime().Format(time.RFC3339)
	}
	if proj.DeleteTime != nil {
		data.DeleteTime = proj.DeleteTime.AsTime().Format(time.RFC3339)
	}

	// Convert labels
	data.Labels = ConvertLabels(proj.Labels)

	return data
}

// ConvertLabels converts project labels from GCP API to label data.
func ConvertLabels(labels map[string]string) []LabelData {
	if len(labels) == 0 {
		return nil
	}

	result := make([]LabelData, 0, len(labels))
	for key, value := range labels {
		result = append(result, LabelData{
			Key:   key,
			Value: value,
		})
	}

	return result
}


// extractProjectNumber extracts the project number from the name field.
// Name format: "projects/123456789" -> "123456789"
func extractProjectNumber(name string) string {
	parts := strings.Split(name, "/")
	if len(parts) == 2 && parts[0] == "projects" {
		return parts[1]
	}
	return name
}
