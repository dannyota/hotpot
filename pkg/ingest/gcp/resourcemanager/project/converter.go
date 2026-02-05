package project

import (
	"strings"
	"time"

	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"

	"hotpot/pkg/base/models/bronze"
)

// ConvertProject converts a GCP API Project to a Bronze model.
// Preserves raw API data with minimal transformation.
func ConvertProject(proj *resourcemanagerpb.Project, collectedAt time.Time) bronze.GCPProject {
	project := bronze.GCPProject{
		ProjectID:     proj.GetProjectId(),
		ProjectNumber: extractProjectNumber(proj.GetName()),
		DisplayName:   proj.GetDisplayName(),
		State:         proj.GetState().String(),
		Parent:        proj.GetParent(),
		Etag:          proj.GetEtag(),
		CollectedAt:   collectedAt,
	}

	// Convert timestamps
	if proj.CreateTime != nil {
		project.CreateTime = proj.CreateTime.AsTime().Format(time.RFC3339)
	}
	if proj.UpdateTime != nil {
		project.UpdateTime = proj.UpdateTime.AsTime().Format(time.RFC3339)
	}
	if proj.DeleteTime != nil {
		project.DeleteTime = proj.DeleteTime.AsTime().Format(time.RFC3339)
	}

	// Convert labels
	project.Labels = ConvertLabels(proj.Labels)

	return project
}

// ConvertLabels converts project labels from GCP API to Bronze models.
func ConvertLabels(labels map[string]string) []bronze.GCPProjectLabel {
	if len(labels) == 0 {
		return nil
	}

	result := make([]bronze.GCPProjectLabel, 0, len(labels))
	for key, value := range labels {
		result = append(result, bronze.GCPProjectLabel{
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
