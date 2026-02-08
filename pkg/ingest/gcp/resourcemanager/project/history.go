package project

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorygcpproject"
	"hotpot/pkg/storage/ent/bronzehistorygcpprojectlabel"
)

// HistoryService manages project history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new project.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, projectData *ProjectData, now time.Time) error {
	// Create project history
	projectHistory, err := tx.BronzeHistoryGCPProject.Create().
		SetProjectID(projectData.ID).
		SetValidFrom(now).
		SetCollectedAt(projectData.CollectedAt).
		SetProjectNumber(projectData.ProjectNumber).
		SetDisplayName(projectData.DisplayName).
		SetState(projectData.State).
		SetParent(projectData.Parent).
		SetCreateTime(projectData.CreateTime).
		SetUpdateTime(projectData.UpdateTime).
		SetDeleteTime(projectData.DeleteTime).
		SetEtag(projectData.Etag).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create project history: %w", err)
	}

	// Create label history
	for _, label := range projectData.Labels {
		_, err := tx.BronzeHistoryGCPProjectLabel.Create().
			SetProjectHistoryID(projectHistory.HistoryID).
			SetValidFrom(now).
			SetKey(label.Key).
			SetValue(label.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label history: %w", err)
		}
	}

	return nil
}

// UpdateHistory updates history records for a changed project.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPProject, new *ProjectData, diff *ProjectDiff, now time.Time) error {
	// Get current project history
	currentHistory, err := tx.BronzeHistoryGCPProject.Query().
		Where(
			bronzehistorygcpproject.ProjectID(old.ID),
			bronzehistorygcpproject.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current project history: %w", err)
	}

	// Close current project history if core fields changed
	if diff.IsChanged {
		// Close old label history first
		_, err := tx.BronzeHistoryGCPProjectLabel.Update().
			Where(
				bronzehistorygcpprojectlabel.ProjectHistoryID(currentHistory.HistoryID),
				bronzehistorygcpprojectlabel.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to close old label history: %w", err)
		}

		// Close current project history
		err = tx.BronzeHistoryGCPProject.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current project history: %w", err)
		}

		// Create new project history
		newHistory, err := tx.BronzeHistoryGCPProject.Create().
			SetProjectID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetProjectNumber(new.ProjectNumber).
			SetDisplayName(new.DisplayName).
			SetState(new.State).
			SetParent(new.Parent).
			SetCreateTime(new.CreateTime).
			SetUpdateTime(new.UpdateTime).
			SetDeleteTime(new.DeleteTime).
			SetEtag(new.Etag).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new project history: %w", err)
		}

		// Create new label history linked to new project history
		for _, label := range new.Labels {
			_, err := tx.BronzeHistoryGCPProjectLabel.Create().
				SetProjectHistoryID(newHistory.HistoryID).
				SetValidFrom(now).
				SetKey(label.Key).
				SetValue(label.Value).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create label history: %w", err)
			}
		}
	} else if diff.LabelsDiff.HasChanges {
		// Only labels changed - close old label history and create new ones
		_, err := tx.BronzeHistoryGCPProjectLabel.Update().
			Where(
				bronzehistorygcpprojectlabel.ProjectHistoryID(currentHistory.HistoryID),
				bronzehistorygcpprojectlabel.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to close label history: %w", err)
		}

		for _, label := range new.Labels {
			_, err := tx.BronzeHistoryGCPProjectLabel.Create().
				SetProjectHistoryID(currentHistory.HistoryID).
				SetValidFrom(now).
				SetKey(label.Key).
				SetValue(label.Value).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create label history: %w", err)
			}
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted project.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, projectID string, now time.Time) error {
	// Get current project history
	currentHistory, err := tx.BronzeHistoryGCPProject.Query().
		Where(
			bronzehistorygcpproject.ProjectID(projectID),
			bronzehistorygcpproject.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("failed to find current project history: %w", err)
	}

	// Close project history
	err = tx.BronzeHistoryGCPProject.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close project history: %w", err)
	}

	// Close label history
	_, err = tx.BronzeHistoryGCPProjectLabel.Update().
		Where(
			bronzehistorygcpprojectlabel.ProjectHistoryID(currentHistory.HistoryID),
			bronzehistorygcpprojectlabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close label history: %w", err)
	}

	return nil
}
