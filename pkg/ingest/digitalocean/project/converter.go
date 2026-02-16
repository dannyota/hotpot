package project

import (
	"fmt"
	"time"

	"github.com/digitalocean/godo"
)

// ProjectData holds converted Project data ready for Ent insertion.
type ProjectData struct {
	ResourceID   string
	OwnerUUID    string
	OwnerID      uint64
	Name         string
	Description  string
	Purpose      string
	Environment  string
	IsDefault    bool
	APICreatedAt string
	APIUpdatedAt string
	CollectedAt  time.Time
}

// ConvertProject converts a godo Project to ProjectData.
func ConvertProject(v godo.Project, collectedAt time.Time) *ProjectData {
	return &ProjectData{
		ResourceID:   v.ID,
		OwnerUUID:    v.OwnerUUID,
		OwnerID:      v.OwnerID,
		Name:         v.Name,
		Description:  v.Description,
		Purpose:      v.Purpose,
		Environment:  v.Environment,
		IsDefault:    v.IsDefault,
		APICreatedAt: v.CreatedAt,
		APIUpdatedAt: v.UpdatedAt,
		CollectedAt:  collectedAt,
	}
}

// ProjectResourceData holds converted Project Resource data ready for Ent insertion.
type ProjectResourceData struct {
	ResourceID  string
	ProjectID   string
	URN         string
	AssignedAt  string
	Status      string
	CollectedAt time.Time
}

// ConvertProjectResource converts a godo ProjectResource to ProjectResourceData.
func ConvertProjectResource(v godo.ProjectResource, projectID string, collectedAt time.Time) *ProjectResourceData {
	return &ProjectResourceData{
		ResourceID:  fmt.Sprintf("%s:%s", projectID, v.URN),
		ProjectID:   projectID,
		URN:         v.URN,
		AssignedAt:  v.AssignedAt,
		Status:      v.Status,
		CollectedAt: collectedAt,
	}
}
