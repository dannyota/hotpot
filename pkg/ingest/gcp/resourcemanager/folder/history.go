package folder

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpfolder"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpfolderlabel"
)

// HistoryService manages folder history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new folder.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, folderData *FolderData, now time.Time) error {
	// Create folder history
	folderHistory, err := tx.BronzeHistoryGCPFolder.Create().
		SetResourceID(folderData.ID).
		SetValidFrom(now).
		SetCollectedAt(folderData.CollectedAt).
		SetFirstCollectedAt(folderData.CollectedAt).
		SetName(folderData.Name).
		SetDisplayName(folderData.DisplayName).
		SetState(folderData.State).
		SetParent(folderData.Parent).
		SetCreateTime(folderData.CreateTime).
		SetUpdateTime(folderData.UpdateTime).
		SetDeleteTime(folderData.DeleteTime).
		SetEtag(folderData.Etag).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create folder history: %w", err)
	}

	// Create label history
	for _, label := range folderData.Labels {
		_, err := tx.BronzeHistoryGCPFolderLabel.Create().
			SetFolderHistoryID(folderHistory.HistoryID).
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

// UpdateHistory updates history records for a changed folder.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPFolder, new *FolderData, diff *FolderDiff, now time.Time) error {
	// Get current folder history
	currentHistory, err := tx.BronzeHistoryGCPFolder.Query().
		Where(
			bronzehistorygcpfolder.ResourceID(old.ID),
			bronzehistorygcpfolder.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current folder history: %w", err)
	}

	// Close current folder history if core fields changed
	if diff.IsChanged {
		// Close old label history first
		_, err := tx.BronzeHistoryGCPFolderLabel.Update().
			Where(
				bronzehistorygcpfolderlabel.FolderHistoryID(currentHistory.HistoryID),
				bronzehistorygcpfolderlabel.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to close old label history: %w", err)
		}

		// Close current folder history
		err = tx.BronzeHistoryGCPFolder.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current folder history: %w", err)
		}

		// Create new folder history
		newHistory, err := tx.BronzeHistoryGCPFolder.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetDisplayName(new.DisplayName).
			SetState(new.State).
			SetParent(new.Parent).
			SetCreateTime(new.CreateTime).
			SetUpdateTime(new.UpdateTime).
			SetDeleteTime(new.DeleteTime).
			SetEtag(new.Etag).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new folder history: %w", err)
		}

		// Create new label history linked to new folder history
		for _, label := range new.Labels {
			_, err := tx.BronzeHistoryGCPFolderLabel.Create().
				SetFolderHistoryID(newHistory.HistoryID).
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
		_, err := tx.BronzeHistoryGCPFolderLabel.Update().
			Where(
				bronzehistorygcpfolderlabel.FolderHistoryID(currentHistory.HistoryID),
				bronzehistorygcpfolderlabel.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to close label history: %w", err)
		}

		for _, label := range new.Labels {
			_, err := tx.BronzeHistoryGCPFolderLabel.Create().
				SetFolderHistoryID(currentHistory.HistoryID).
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

// CloseHistory closes all history records for a deleted folder.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, folderID string, now time.Time) error {
	// Get current folder history
	currentHistory, err := tx.BronzeHistoryGCPFolder.Query().
		Where(
			bronzehistorygcpfolder.ResourceID(folderID),
			bronzehistorygcpfolder.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("failed to find current folder history: %w", err)
	}

	// Close folder history
	err = tx.BronzeHistoryGCPFolder.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close folder history: %w", err)
	}

	// Close label history
	_, err = tx.BronzeHistoryGCPFolderLabel.Update().
		Where(
			bronzehistorygcpfolderlabel.FolderHistoryID(currentHistory.HistoryID),
			bronzehistorygcpfolderlabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close label history: %w", err)
	}

	return nil
}
