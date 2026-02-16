package project

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzedoproject"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzedoprojectresource"
)

// Service handles DigitalOcean Project ingestion.
type Service struct {
	client          *Client
	entClient       *ent.Client
	projectHistory  *ProjectHistoryService
	resourceHistory *ResourceHistoryService
}

// NewService creates a new Project ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:          client,
		entClient:       entClient,
		projectHistory:  NewProjectHistoryService(entClient),
		resourceHistory: NewResourceHistoryService(entClient),
	}
}

// IngestProjectsResult contains the result of Project ingestion.
type IngestProjectsResult struct {
	ProjectCount   int
	CollectedAt    time.Time
	DurationMillis int64
	ProjectIDs     []string
}

// IngestProjects fetches all Projects from DigitalOcean and saves them.
func (s *Service) IngestProjects(ctx context.Context, heartbeat func()) (*IngestProjectsResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	apiProjects, err := s.client.ListAllProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	if heartbeat != nil {
		heartbeat()
	}

	var allProjects []*ProjectData
	var projectIDs []string
	for _, v := range apiProjects {
		allProjects = append(allProjects, ConvertProject(v, collectedAt))
		projectIDs = append(projectIDs, v.ID)
	}

	if err := s.saveProjects(ctx, allProjects); err != nil {
		return nil, fmt.Errorf("save projects: %w", err)
	}

	return &IngestProjectsResult{
		ProjectCount:   len(allProjects),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
		ProjectIDs:     projectIDs,
	}, nil
}

func (s *Service) saveProjects(ctx context.Context, projects []*ProjectData) error {
	if len(projects) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, data := range projects {
		existing, err := tx.BronzeDOProject.Query().
			Where(bronzedoproject.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing Project %s: %w", data.ResourceID, err)
		}

		diff := DiffProjectData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeDOProject.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for Project %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			if _, err := tx.BronzeDOProject.Create().
				SetID(data.ResourceID).
				SetOwnerUUID(data.OwnerUUID).
				SetOwnerID(data.OwnerID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetPurpose(data.Purpose).
				SetEnvironment(data.Environment).
				SetIsDefault(data.IsDefault).
				SetAPICreatedAt(data.APICreatedAt).
				SetAPIUpdatedAt(data.APIUpdatedAt).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create Project %s: %w", data.ResourceID, err)
			}

			if err := s.projectHistory.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for Project %s: %w", data.ResourceID, err)
			}
		} else {
			if _, err := tx.BronzeDOProject.UpdateOneID(data.ResourceID).
				SetOwnerUUID(data.OwnerUUID).
				SetOwnerID(data.OwnerID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetPurpose(data.Purpose).
				SetEnvironment(data.Environment).
				SetIsDefault(data.IsDefault).
				SetAPICreatedAt(data.APICreatedAt).
				SetAPIUpdatedAt(data.APIUpdatedAt).
				SetCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update Project %s: %w", data.ResourceID, err)
			}

			if err := s.projectHistory.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for Project %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleProjects removes Projects that were not collected in the latest run.
func (s *Service) DeleteStaleProjects(ctx context.Context, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	stale, err := tx.BronzeDOProject.Query().
		Where(bronzedoproject.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, p := range stale {
		if err := s.projectHistory.CloseHistory(ctx, tx, p.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for Project %s: %w", p.ID, err)
		}

		if err := tx.BronzeDOProject.DeleteOne(p).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete Project %s: %w", p.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// IngestResourcesResult contains the result of Project Resource ingestion.
type IngestResourcesResult struct {
	ResourceCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// IngestResources fetches all Project Resources for given projects and saves them.
func (s *Service) IngestResources(ctx context.Context, projectIDs []string, heartbeat func()) (*IngestResourcesResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	var allResources []*ProjectResourceData
	for _, projectID := range projectIDs {
		apiResources, err := s.client.ListAllResources(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("list resources for project %s: %w", projectID, err)
		}

		for _, v := range apiResources {
			allResources = append(allResources, ConvertProjectResource(v, projectID, collectedAt))
		}

		if heartbeat != nil {
			heartbeat()
		}
	}

	if err := s.saveResources(ctx, allResources); err != nil {
		return nil, fmt.Errorf("save project resources: %w", err)
	}

	return &IngestResourcesResult{
		ResourceCount:  len(allResources),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveResources(ctx context.Context, resources []*ProjectResourceData) error {
	if len(resources) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, data := range resources {
		existing, err := tx.BronzeDOProjectResource.Query().
			Where(bronzedoprojectresource.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing ProjectResource %s: %w", data.ResourceID, err)
		}

		diff := DiffProjectResourceData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeDOProjectResource.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for ProjectResource %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			if _, err := tx.BronzeDOProjectResource.Create().
				SetID(data.ResourceID).
				SetProjectID(data.ProjectID).
				SetUrn(data.URN).
				SetAssignedAt(data.AssignedAt).
				SetStatus(data.Status).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create ProjectResource %s: %w", data.ResourceID, err)
			}

			if err := s.resourceHistory.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for ProjectResource %s: %w", data.ResourceID, err)
			}
		} else {
			if _, err := tx.BronzeDOProjectResource.UpdateOneID(data.ResourceID).
				SetProjectID(data.ProjectID).
				SetUrn(data.URN).
				SetAssignedAt(data.AssignedAt).
				SetStatus(data.Status).
				SetCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update ProjectResource %s: %w", data.ResourceID, err)
			}

			if err := s.resourceHistory.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for ProjectResource %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleResources removes Project Resources that were not collected in the latest run.
func (s *Service) DeleteStaleResources(ctx context.Context, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	stale, err := tx.BronzeDOProjectResource.Query().
		Where(bronzedoprojectresource.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, r := range stale {
		if err := s.resourceHistory.CloseHistory(ctx, tx, r.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for ProjectResource %s: %w", r.ID, err)
		}

		if err := tx.BronzeDOProjectResource.DeleteOne(r).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete ProjectResource %s: %w", r.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
