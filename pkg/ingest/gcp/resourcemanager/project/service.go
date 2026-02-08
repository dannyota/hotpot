package project

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzegcpproject"
	"hotpot/pkg/storage/ent/bronzegcpprojectlabel"
)

// Service handles GCP Project ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new project ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of project ingestion.
type IngestResult struct {
	ProjectCount   int
	ProjectIDs     []string
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches all accessible projects from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch projects from GCP
	projects, err := s.client.SearchProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to search projects: %w", err)
	}

	// Convert to project data
	projectDataList := make([]*ProjectData, 0, len(projects))
	projectIDs := make([]string, 0, len(projects))
	for _, proj := range projects {
		data := ConvertProject(proj, collectedAt)
		projectDataList = append(projectDataList, data)
		projectIDs = append(projectIDs, data.ID)
	}

	// Save to database
	if err := s.saveProjects(ctx, projectDataList); err != nil {
		return nil, fmt.Errorf("failed to save projects: %w", err)
	}

	return &IngestResult{
		ProjectCount:   len(projectDataList),
		ProjectIDs:     projectIDs,
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveProjects saves projects to the database with history tracking.
func (s *Service) saveProjects(ctx context.Context, projects []*ProjectData) error {
	if len(projects) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, projectData := range projects {
		// Load existing project with labels
		existing, err := tx.BronzeGCPProject.Query().
			Where(bronzegcpproject.ID(projectData.ID)).
			WithLabels().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing project %s: %w", projectData.ID, err)
		}

		// Compute diff
		diff := DiffProjectData(existing, projectData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPProject.UpdateOneID(projectData.ID).
				SetCollectedAt(projectData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for project %s: %w", projectData.ID, err)
			}
			continue
		}

		// Delete old labels if updating
		if existing != nil {
			_, err := tx.BronzeGCPProjectLabel.Delete().
				Where(bronzegcpprojectlabel.HasProjectWith(bronzegcpproject.ID(projectData.ID))).
				Exec(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old labels for project %s: %w", projectData.ID, err)
			}
		}

		// Create or update project
		var savedProject *ent.BronzeGCPProject
		if existing == nil {
			// Create new project
			create := tx.BronzeGCPProject.Create().
				SetID(projectData.ID).
				SetProjectNumber(projectData.ProjectNumber).
				SetState(projectData.State).
				SetEtag(projectData.Etag).
				SetCollectedAt(projectData.CollectedAt).
				SetFirstCollectedAt(projectData.CollectedAt)

			if projectData.DisplayName != "" {
				create.SetDisplayName(projectData.DisplayName)
			}
			if projectData.Parent != "" {
				create.SetParent(projectData.Parent)
			}
			if projectData.CreateTime != "" {
				create.SetCreateTime(projectData.CreateTime)
			}
			if projectData.UpdateTime != "" {
				create.SetUpdateTime(projectData.UpdateTime)
			}
			if projectData.DeleteTime != "" {
				create.SetDeleteTime(projectData.DeleteTime)
			}

			savedProject, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create project %s: %w", projectData.ID, err)
			}

			// Create labels for new project
			for _, label := range projectData.Labels {
				_, err := tx.BronzeGCPProjectLabel.Create().
					SetKey(label.Key).
					SetValue(label.Value).
					SetProject(savedProject).
					Save(ctx)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to create label for project %s: %w", projectData.ID, err)
				}
			}
		} else {
			// Update existing project
			update := tx.BronzeGCPProject.UpdateOneID(projectData.ID).
				SetProjectNumber(projectData.ProjectNumber).
				SetState(projectData.State).
				SetEtag(projectData.Etag).
				SetCollectedAt(projectData.CollectedAt)

			if projectData.DisplayName != "" {
				update.SetDisplayName(projectData.DisplayName)
			}
			if projectData.Parent != "" {
				update.SetParent(projectData.Parent)
			}
			if projectData.CreateTime != "" {
				update.SetCreateTime(projectData.CreateTime)
			}
			if projectData.UpdateTime != "" {
				update.SetUpdateTime(projectData.UpdateTime)
			}
			if projectData.DeleteTime != "" {
				update.SetDeleteTime(projectData.DeleteTime)
			}

			savedProject, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update project %s: %w", projectData.ID, err)
			}

			// Create new labels
			for _, label := range projectData.Labels {
				_, err := tx.BronzeGCPProjectLabel.Create().
					SetKey(label.Key).
					SetValue(label.Value).
					SetProject(savedProject).
					Save(ctx)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to create label for project %s: %w", projectData.ID, err)
				}
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, projectData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for project %s: %w", projectData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, projectData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for project %s: %w", projectData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleProjects removes projects that were not collected in the latest run.
// Also closes history records for deleted projects.
func (s *Service) DeleteStaleProjects(ctx context.Context, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	// Find stale projects
	staleProjects, err := tx.BronzeGCPProject.Query().
		Where(bronzegcpproject.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale project
	for _, proj := range staleProjects {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, proj.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for project %s: %w", proj.ID, err)
		}

		// Delete project (labels will be deleted automatically via CASCADE)
		if err := tx.BronzeGCPProject.DeleteOne(proj).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete project %s: %w", proj.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
