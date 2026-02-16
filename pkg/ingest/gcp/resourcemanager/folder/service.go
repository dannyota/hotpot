package folder

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpfolder"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpfolderlabel"
)

// Service handles GCP Folder ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new folder ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of folder ingestion.
type IngestResult struct {
	FolderCount    int
	FolderIDs      []string
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches all accessible folders from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch folders from GCP
	folders, err := s.client.SearchFolders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to search folders: %w", err)
	}

	// Convert to folder data
	folderDataList := make([]*FolderData, 0, len(folders))
	folderIDs := make([]string, 0, len(folders))
	for _, f := range folders {
		data := ConvertFolder(f, collectedAt)
		folderDataList = append(folderDataList, data)
		folderIDs = append(folderIDs, data.ID)
	}

	// Save to database
	if err := s.saveFolders(ctx, folderDataList); err != nil {
		return nil, fmt.Errorf("failed to save folders: %w", err)
	}

	return &IngestResult{
		FolderCount:    len(folderDataList),
		FolderIDs:      folderIDs,
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveFolders saves folders to the database with history tracking.
func (s *Service) saveFolders(ctx context.Context, folders []*FolderData) error {
	if len(folders) == 0 {
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

	for _, folderData := range folders {
		// Load existing folder with labels
		existing, err := tx.BronzeGCPFolder.Query().
			Where(bronzegcpfolder.ID(folderData.ID)).
			WithLabels().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing folder %s: %w", folderData.ID, err)
		}

		// Compute diff
		diff := DiffFolderData(existing, folderData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPFolder.UpdateOneID(folderData.ID).
				SetCollectedAt(folderData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for folder %s: %w", folderData.ID, err)
			}
			continue
		}

		// Delete old labels if updating
		if existing != nil {
			_, err := tx.BronzeGCPFolderLabel.Delete().
				Where(bronzegcpfolderlabel.HasFolderWith(bronzegcpfolder.ID(folderData.ID))).
				Exec(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old labels for folder %s: %w", folderData.ID, err)
			}
		}

		// Create or update folder
		var savedFolder *ent.BronzeGCPFolder
		if existing == nil {
			// Create new folder
			create := tx.BronzeGCPFolder.Create().
				SetID(folderData.ID).
				SetName(folderData.Name).
				SetState(folderData.State).
				SetEtag(folderData.Etag).
				SetCollectedAt(folderData.CollectedAt).
				SetFirstCollectedAt(folderData.CollectedAt)

			if folderData.DisplayName != "" {
				create.SetDisplayName(folderData.DisplayName)
			}
			if folderData.Parent != "" {
				create.SetParent(folderData.Parent)
			}
			if folderData.CreateTime != "" {
				create.SetCreateTime(folderData.CreateTime)
			}
			if folderData.UpdateTime != "" {
				create.SetUpdateTime(folderData.UpdateTime)
			}
			if folderData.DeleteTime != "" {
				create.SetDeleteTime(folderData.DeleteTime)
			}

			savedFolder, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create folder %s: %w", folderData.ID, err)
			}

			// Create labels for new folder
			for _, label := range folderData.Labels {
				_, err := tx.BronzeGCPFolderLabel.Create().
					SetKey(label.Key).
					SetValue(label.Value).
					SetFolder(savedFolder).
					Save(ctx)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to create label for folder %s: %w", folderData.ID, err)
				}
			}
		} else {
			// Update existing folder
			update := tx.BronzeGCPFolder.UpdateOneID(folderData.ID).
				SetName(folderData.Name).
				SetState(folderData.State).
				SetEtag(folderData.Etag).
				SetCollectedAt(folderData.CollectedAt)

			if folderData.DisplayName != "" {
				update.SetDisplayName(folderData.DisplayName)
			}
			if folderData.Parent != "" {
				update.SetParent(folderData.Parent)
			}
			if folderData.CreateTime != "" {
				update.SetCreateTime(folderData.CreateTime)
			}
			if folderData.UpdateTime != "" {
				update.SetUpdateTime(folderData.UpdateTime)
			}
			if folderData.DeleteTime != "" {
				update.SetDeleteTime(folderData.DeleteTime)
			}

			savedFolder, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update folder %s: %w", folderData.ID, err)
			}

			// Create new labels
			for _, label := range folderData.Labels {
				_, err := tx.BronzeGCPFolderLabel.Create().
					SetKey(label.Key).
					SetValue(label.Value).
					SetFolder(savedFolder).
					Save(ctx)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to create label for folder %s: %w", folderData.ID, err)
				}
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, folderData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for folder %s: %w", folderData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, folderData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for folder %s: %w", folderData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleFolders removes folders that were not collected in the latest run.
// Also closes history records for deleted folders.
func (s *Service) DeleteStaleFolders(ctx context.Context, collectedAt time.Time) error {
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

	// Find stale folders
	staleFolders, err := tx.BronzeGCPFolder.Query().
		Where(bronzegcpfolder.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale folder
	for _, f := range staleFolders {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, f.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for folder %s: %w", f.ID, err)
		}

		// Delete folder (labels will be deleted automatically via CASCADE)
		if err := tx.BronzeGCPFolder.DeleteOne(f).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete folder %s: %w", f.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
