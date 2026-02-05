package project

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
	"hotpot/pkg/base/models/bronze_history"
)

// HistoryService manages project history tracking.
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService creates a new history service.
func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

// CreateHistory creates initial history records for a new project.
func (h *HistoryService) CreateHistory(tx *gorm.DB, project *bronze.GCPProject, now time.Time) error {
	// Create project history
	projectHistory := bronze_history.GCPProject{
		ProjectID:     project.ProjectID,
		ValidFrom:     now,
		ValidTo:       nil,
		ProjectNumber: project.ProjectNumber,
		DisplayName:   project.DisplayName,
		State:         project.State,
		Parent:        project.Parent,
		CreateTime:    project.CreateTime,
		UpdateTime:    project.UpdateTime,
		DeleteTime:    project.DeleteTime,
		Etag:          project.Etag,
		CollectedAt:   project.CollectedAt,
	}

	if err := tx.Create(&projectHistory).Error; err != nil {
		return fmt.Errorf("failed to create project history: %w", err)
	}

	// Create label history
	for _, label := range project.Labels {
		labelHistory := bronze_history.GCPProjectLabel{
			ProjectHistoryID: projectHistory.HistoryID,
			ValidFrom:        now,
			ValidTo:          nil,
			Key:              label.Key,
			Value:            label.Value,
		}
		if err := tx.Create(&labelHistory).Error; err != nil {
			return fmt.Errorf("failed to create label history: %w", err)
		}
	}

	return nil
}

// UpdateHistory updates history records for a changed project.
func (h *HistoryService) UpdateHistory(tx *gorm.DB, old, new *bronze.GCPProject, diff *ProjectDiff, now time.Time) error {
	// Get current project history
	var currentHistory bronze_history.GCPProject
	if err := tx.Where("project_id = ? AND valid_to IS NULL", old.ProjectID).
		First(&currentHistory).Error; err != nil {
		return fmt.Errorf("failed to find current project history: %w", err)
	}

	// Close current project history if core fields changed
	if diff.IsChanged {
		// Close old label history first
		if err := tx.Model(&bronze_history.GCPProjectLabel{}).
			Where("project_history_id = ? AND valid_to IS NULL", currentHistory.HistoryID).
			Update("valid_to", now).Error; err != nil {
			return fmt.Errorf("failed to close old label history: %w", err)
		}

		if err := tx.Model(&currentHistory).Update("valid_to", now).Error; err != nil {
			return fmt.Errorf("failed to close current project history: %w", err)
		}

		// Create new project history
		newHistory := bronze_history.GCPProject{
			ProjectID:     new.ProjectID,
			ValidFrom:     now,
			ValidTo:       nil,
			ProjectNumber: new.ProjectNumber,
			DisplayName:   new.DisplayName,
			State:         new.State,
			Parent:        new.Parent,
			CreateTime:    new.CreateTime,
			UpdateTime:    new.UpdateTime,
			DeleteTime:    new.DeleteTime,
			Etag:          new.Etag,
			CollectedAt:   new.CollectedAt,
		}
		if err := tx.Create(&newHistory).Error; err != nil {
			return fmt.Errorf("failed to create new project history: %w", err)
		}

		// Create new label history linked to new project history
		for _, label := range new.Labels {
			labelHistory := bronze_history.GCPProjectLabel{
				ProjectHistoryID: newHistory.HistoryID,
				ValidFrom:        now,
				ValidTo:          nil,
				Key:              label.Key,
				Value:            label.Value,
			}
			if err := tx.Create(&labelHistory).Error; err != nil {
				return fmt.Errorf("failed to create label history: %w", err)
			}
		}
	} else if diff.LabelsDiff.HasChanges {
		// Only labels changed - close old label history and create new ones
		if err := tx.Model(&bronze_history.GCPProjectLabel{}).
			Where("project_history_id = ? AND valid_to IS NULL", currentHistory.HistoryID).
			Update("valid_to", now).Error; err != nil {
			return fmt.Errorf("failed to close label history: %w", err)
		}

		for _, label := range new.Labels {
			labelHistory := bronze_history.GCPProjectLabel{
				ProjectHistoryID: currentHistory.HistoryID,
				ValidFrom:        now,
				ValidTo:          nil,
				Key:              label.Key,
				Value:            label.Value,
			}
			if err := tx.Create(&labelHistory).Error; err != nil {
				return fmt.Errorf("failed to create label history: %w", err)
			}
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted project.
func (h *HistoryService) CloseHistory(tx *gorm.DB, projectID string, now time.Time) error {
	// Get current project history
	var currentHistory bronze_history.GCPProject
	if err := tx.Where("project_id = ? AND valid_to IS NULL", projectID).
		First(&currentHistory).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // No history to close
		}
		return fmt.Errorf("failed to find current project history: %w", err)
	}

	// Close project history
	if err := tx.Model(&currentHistory).Update("valid_to", now).Error; err != nil {
		return fmt.Errorf("failed to close project history: %w", err)
	}

	// Close label history
	if err := tx.Model(&bronze_history.GCPProjectLabel{}).
		Where("project_history_id = ? AND valid_to IS NULL", currentHistory.HistoryID).
		Update("valid_to", now).Error; err != nil {
		return fmt.Errorf("failed to close label history: %w", err)
	}

	return nil
}
