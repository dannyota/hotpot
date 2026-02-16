package project

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorydoproject"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorydoprojectresource"
)

// ProjectHistoryService handles history tracking for Projects.
type ProjectHistoryService struct {
	entClient *ent.Client
}

// NewProjectHistoryService creates a new project history service.
func NewProjectHistoryService(entClient *ent.Client) *ProjectHistoryService {
	return &ProjectHistoryService{entClient: entClient}
}

func (h *ProjectHistoryService) buildCreate(tx *ent.Tx, data *ProjectData) *ent.BronzeHistoryDOProjectCreate {
	return tx.BronzeHistoryDOProject.Create().
		SetResourceID(data.ResourceID).
		SetOwnerUUID(data.OwnerUUID).
		SetOwnerID(data.OwnerID).
		SetName(data.Name).
		SetDescription(data.Description).
		SetPurpose(data.Purpose).
		SetEnvironment(data.Environment).
		SetIsDefault(data.IsDefault).
		SetAPICreatedAt(data.APICreatedAt).
		SetAPIUpdatedAt(data.APIUpdatedAt)
}

// CreateHistory creates a history record for a new Project.
func (h *ProjectHistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *ProjectData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create Project history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new for a changed Project.
func (h *ProjectHistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeDOProject, new *ProjectData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDOProject.Query().
		Where(
			bronzehistorydoproject.ResourceID(old.ID),
			bronzehistorydoproject.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current Project history: %w", err)
	}

	if err := tx.BronzeHistoryDOProject.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close Project history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new Project history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted Project.
func (h *ProjectHistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDOProject.Query().
		Where(
			bronzehistorydoproject.ResourceID(resourceID),
			bronzehistorydoproject.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current Project history: %w", err)
	}

	if err := tx.BronzeHistoryDOProject.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close Project history: %w", err)
	}

	return nil
}

// ResourceHistoryService handles history tracking for Project Resources.
type ResourceHistoryService struct {
	entClient *ent.Client
}

// NewResourceHistoryService creates a new project resource history service.
func NewResourceHistoryService(entClient *ent.Client) *ResourceHistoryService {
	return &ResourceHistoryService{entClient: entClient}
}

func (h *ResourceHistoryService) buildCreate(tx *ent.Tx, data *ProjectResourceData) *ent.BronzeHistoryDOProjectResourceCreate {
	return tx.BronzeHistoryDOProjectResource.Create().
		SetResourceID(data.ResourceID).
		SetProjectID(data.ProjectID).
		SetUrn(data.URN).
		SetAssignedAt(data.AssignedAt).
		SetStatus(data.Status)
}

// CreateHistory creates a history record for a new Project Resource.
func (h *ResourceHistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *ProjectResourceData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create ProjectResource history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new for a changed Project Resource.
func (h *ResourceHistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeDOProjectResource, new *ProjectResourceData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDOProjectResource.Query().
		Where(
			bronzehistorydoprojectresource.ResourceID(old.ID),
			bronzehistorydoprojectresource.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current ProjectResource history: %w", err)
	}

	if err := tx.BronzeHistoryDOProjectResource.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close ProjectResource history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new ProjectResource history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted Project Resource.
func (h *ResourceHistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDOProjectResource.Query().
		Where(
			bronzehistorydoprojectresource.ResourceID(resourceID),
			bronzehistorydoprojectresource.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current ProjectResource history: %w", err)
	}

	if err := tx.BronzeHistoryDOProjectResource.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close ProjectResource history: %w", err)
	}

	return nil
}
