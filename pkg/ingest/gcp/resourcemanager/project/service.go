package project

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
)

// Service handles GCP Project ingestion.
type Service struct {
	client  *Client
	db      *gorm.DB
	history *HistoryService
}

// NewService creates a new project ingestion service.
func NewService(client *Client, db *gorm.DB) *Service {
	return &Service{
		client:  client,
		db:      db,
		history: NewHistoryService(db),
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

	// Convert to bronze models
	bronzeProjects := make([]bronze.GCPProject, 0, len(projects))
	projectIDs := make([]string, 0, len(projects))
	for _, proj := range projects {
		bronzeProj := ConvertProject(proj, collectedAt)
		bronzeProjects = append(bronzeProjects, bronzeProj)
		projectIDs = append(projectIDs, bronzeProj.ProjectID)
	}

	// Save to database
	if err := s.saveProjects(ctx, bronzeProjects); err != nil {
		return nil, fmt.Errorf("failed to save projects: %w", err)
	}

	return &IngestResult{
		ProjectCount:   len(bronzeProjects),
		ProjectIDs:     projectIDs,
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveProjects saves projects to the database with history tracking.
func (s *Service) saveProjects(ctx context.Context, projects []bronze.GCPProject) error {
	if len(projects) == 0 {
		return nil
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, project := range projects {
			// Load existing project with labels
			var existing *bronze.GCPProject
			var old bronze.GCPProject
			err := tx.Preload("Labels").
				Where("project_id = ?", project.ProjectID).
				First(&old).Error
			if err == nil {
				existing = &old
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to load existing project %s: %w", project.ProjectID, err)
			}

			// Compute diff
			diff := DiffProject(existing, &project)

			// Skip if no changes
			if !diff.HasAnyChange() && existing != nil {
				// Update collected_at only
				if err := tx.Model(&bronze.GCPProject{}).
					Where("project_id = ?", project.ProjectID).
					Update("collected_at", project.CollectedAt).Error; err != nil {
					return fmt.Errorf("failed to update collected_at for project %s: %w", project.ProjectID, err)
				}
				continue
			}

			// Delete old labels
			if existing != nil {
				if err := tx.Where("project_id = ?", project.ProjectID).
					Delete(&bronze.GCPProjectLabel{}).Error; err != nil {
					return fmt.Errorf("failed to delete old labels for project %s: %w", project.ProjectID, err)
				}
			}

			// Upsert project
			if err := tx.Save(&project).Error; err != nil {
				return fmt.Errorf("failed to upsert project %s: %w", project.ProjectID, err)
			}

			// Create new labels
			for i := range project.Labels {
				project.Labels[i].ProjectID = project.ProjectID
			}
			if len(project.Labels) > 0 {
				if err := tx.Create(&project.Labels).Error; err != nil {
					return fmt.Errorf("failed to create labels for project %s: %w", project.ProjectID, err)
				}
			}

			// Track history
			if diff.IsNew {
				if err := s.history.CreateHistory(tx, &project, now); err != nil {
					return fmt.Errorf("failed to create history for project %s: %w", project.ProjectID, err)
				}
			} else {
				if err := s.history.UpdateHistory(tx, existing, &project, diff, now); err != nil {
					return fmt.Errorf("failed to update history for project %s: %w", project.ProjectID, err)
				}
			}
		}

		return nil
	})
}

// DeleteStaleProjects removes projects that were not collected in the latest run.
// Also closes history records for deleted projects.
func (s *Service) DeleteStaleProjects(ctx context.Context, collectedAt time.Time) error {
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find stale projects
		var staleProjects []bronze.GCPProject
		if err := tx.Where("collected_at < ?", collectedAt).
			Find(&staleProjects).Error; err != nil {
			return err
		}

		// Close history and delete each stale project
		for _, proj := range staleProjects {
			// Close history
			if err := s.history.CloseHistory(tx, proj.ProjectID, now); err != nil {
				return fmt.Errorf("failed to close history for project %s: %w", proj.ProjectID, err)
			}

			// Delete labels
			if err := tx.Where("project_id = ?", proj.ProjectID).
				Delete(&bronze.GCPProjectLabel{}).Error; err != nil {
				return fmt.Errorf("failed to delete labels for project %s: %w", proj.ProjectID, err)
			}

			// Delete project
			if err := tx.Delete(&proj).Error; err != nil {
				return fmt.Errorf("failed to delete project %s: %w", proj.ProjectID, err)
			}
		}

		return nil
	})
}
