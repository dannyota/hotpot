package folder

import (
	"time"

	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
)

// FolderData holds converted folder data ready for Ent insertion.
type FolderData struct {
	ID          string
	Name        string
	DisplayName string
	State       string
	Parent      string
	Etag        string
	CreateTime  string
	UpdateTime  string
	DeleteTime  string
	Labels      []LabelData
	CollectedAt time.Time
}

// LabelData holds converted label data.
type LabelData struct {
	Key   string
	Value string
}

// ConvertFolder converts a GCP API Folder to Ent-compatible data.
// Preserves raw API data with minimal transformation.
func ConvertFolder(folder *resourcemanagerpb.Folder, collectedAt time.Time) *FolderData {
	data := &FolderData{
		ID:          folder.GetName(),
		Name:        folder.GetName(),
		DisplayName: folder.GetDisplayName(),
		State:       folder.GetState().String(),
		Parent:      folder.GetParent(),
		Etag:        folder.GetEtag(),
		CollectedAt: collectedAt,
	}

	// Convert timestamps
	if folder.CreateTime != nil {
		data.CreateTime = folder.CreateTime.AsTime().Format(time.RFC3339)
	}
	if folder.UpdateTime != nil {
		data.UpdateTime = folder.UpdateTime.AsTime().Format(time.RFC3339)
	}
	if folder.DeleteTime != nil {
		data.DeleteTime = folder.DeleteTime.AsTime().Format(time.RFC3339)
	}

	return data
}

// ConvertLabels converts folder labels from GCP API to label data.
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
