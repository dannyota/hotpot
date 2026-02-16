package project

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// ProjectDiff represents changes between old and new Project states.
type ProjectDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffProjectData compares old Ent entity and new data.
func DiffProjectData(old *ent.BronzeDOProject, new *ProjectData) *ProjectDiff {
	if old == nil {
		return &ProjectDiff{IsNew: true}
	}

	changed := old.OwnerUUID != new.OwnerUUID ||
		old.OwnerID != new.OwnerID ||
		old.Name != new.Name ||
		old.Description != new.Description ||
		old.Purpose != new.Purpose ||
		old.Environment != new.Environment ||
		old.IsDefault != new.IsDefault ||
		old.APICreatedAt != new.APICreatedAt ||
		old.APIUpdatedAt != new.APIUpdatedAt

	return &ProjectDiff{IsChanged: changed}
}

// ProjectResourceDiff represents changes between old and new Project Resource states.
type ProjectResourceDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffProjectResourceData compares old Ent entity and new data.
func DiffProjectResourceData(old *ent.BronzeDOProjectResource, new *ProjectResourceData) *ProjectResourceDiff {
	if old == nil {
		return &ProjectResourceDiff{IsNew: true}
	}

	changed := old.ProjectID != new.ProjectID ||
		old.Urn != new.URN ||
		old.AssignedAt != new.AssignedAt ||
		old.Status != new.Status

	return &ProjectResourceDiff{IsChanged: changed}
}
