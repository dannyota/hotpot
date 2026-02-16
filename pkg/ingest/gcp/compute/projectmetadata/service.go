package projectmetadata

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcomputeprojectmetadata"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcomputeprojectmetadataitem"
)

// Service handles GCP Compute project metadata ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new project metadata ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for project metadata ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of project metadata ingestion.
type IngestResult struct {
	ProjectID      string
	MetadataCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches project metadata from GCP and stores it in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch project metadata from GCP (single object per project)
	project, err := s.client.GetProjectMetadata(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project metadata: %w", err)
	}

	// Convert to data struct
	data, err := ConvertProjectMetadata(project, params.ProjectID, collectedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to convert project metadata: %w", err)
	}

	// Save to database (list of 0 or 1 items)
	dataList := []*ProjectMetadataData{data}
	if err := s.saveProjectMetadata(ctx, dataList); err != nil {
		return nil, fmt.Errorf("failed to save project metadata: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		MetadataCount:  1,
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveProjectMetadata saves project metadata to the database with history tracking.
func (s *Service) saveProjectMetadata(ctx context.Context, metadataList []*ProjectMetadataData) error {
	if len(metadataList) == 0 {
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

	for _, data := range metadataList {
		// Load existing metadata with items
		existing, err := tx.BronzeGCPComputeProjectMetadata.Query().
			Where(bronzegcpcomputeprojectmetadata.ID(data.ID)).
			WithItems().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing project metadata %s: %w", data.Name, err)
		}

		// Compute diff
		diff := DiffProjectMetadataData(existing, data)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPComputeProjectMetadata.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for project metadata %s: %w", data.Name, err)
			}
			continue
		}

		// Delete old children if updating
		if existing != nil {
			if err := deleteMetadataChildren(ctx, tx, data.ID); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old children for project metadata %s: %w", data.Name, err)
			}
		}

		// Create or update metadata
		var savedMetadata *ent.BronzeGCPComputeProjectMetadata
		if existing == nil {
			// Create new metadata
			create := tx.BronzeGCPComputeProjectMetadata.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetDefaultServiceAccount(data.DefaultServiceAccount).
				SetDefaultNetworkTier(data.DefaultNetworkTier).
				SetXpnProjectStatus(data.XpnProjectStatus).
				SetCreationTimestamp(data.CreationTimestamp).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.UsageExportLocationJSON != nil {
				create.SetUsageExportLocationJSON(data.UsageExportLocationJSON)
			}

			savedMetadata, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create project metadata %s: %w", data.Name, err)
			}
		} else {
			// Update existing metadata
			update := tx.BronzeGCPComputeProjectMetadata.UpdateOneID(data.ID).
				SetName(data.Name).
				SetDefaultServiceAccount(data.DefaultServiceAccount).
				SetDefaultNetworkTier(data.DefaultNetworkTier).
				SetXpnProjectStatus(data.XpnProjectStatus).
				SetCreationTimestamp(data.CreationTimestamp).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt)

			if data.UsageExportLocationJSON != nil {
				update.SetUsageExportLocationJSON(data.UsageExportLocationJSON)
			}

			savedMetadata, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update project metadata %s: %w", data.Name, err)
			}
		}

		// Create new children
		if err := createMetadataChildren(ctx, tx, savedMetadata, data); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create children for project metadata %s: %w", data.Name, err)
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for project metadata %s: %w", data.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, data, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for project metadata %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// deleteMetadataChildren deletes items for a project metadata record.
func deleteMetadataChildren(ctx context.Context, tx *ent.Tx, metadataID string) error {
	_, err := tx.BronzeGCPComputeProjectMetadataItem.Delete().
		Where(bronzegcpcomputeprojectmetadataitem.HasMetadataWith(bronzegcpcomputeprojectmetadata.ID(metadataID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete items: %w", err)
	}

	return nil
}

// createMetadataChildren creates items for a project metadata record.
func createMetadataChildren(ctx context.Context, tx *ent.Tx, savedMetadata *ent.BronzeGCPComputeProjectMetadata, data *ProjectMetadataData) error {
	for _, item := range data.Items {
		create := tx.BronzeGCPComputeProjectMetadataItem.Create().
			SetKey(item.Key).
			SetMetadata(savedMetadata)
		if item.Value != "" {
			create.SetValue(item.Value)
		}
		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create item: %w", err)
		}
	}

	return nil
}

// DeleteStaleProjectMetadata removes project metadata that was not collected in the latest run.
// Also closes history records for deleted metadata.
func (s *Service) DeleteStaleProjectMetadata(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale project metadata
	staleMetadata, err := tx.BronzeGCPComputeProjectMetadata.Query().
		Where(
			bronzegcpcomputeprojectmetadata.ProjectID(projectID),
			bronzegcpcomputeprojectmetadata.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale record
	for _, m := range staleMetadata {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, m.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for project metadata %s: %w", m.ID, err)
		}

		// Delete metadata (items will be deleted automatically via CASCADE)
		if err := tx.BronzeGCPComputeProjectMetadata.DeleteOne(m).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete project metadata %s: %w", m.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
