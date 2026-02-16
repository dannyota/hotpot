package dataset

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpbigquerydataset"
)

// Service handles BigQuery dataset ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new BigQuery dataset ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for dataset ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of dataset ingestion.
type IngestResult struct {
	ProjectID      string
	DatasetCount   int
	DatasetIDs     []string
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches BigQuery datasets from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch datasets from GCP
	rawDatasets, err := s.client.ListDatasets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list datasets: %w", err)
	}

	// Convert to dataset data
	datasetDataList := make([]*DatasetData, 0, len(rawDatasets))
	for _, raw := range rawDatasets {
		data, err := ConvertDataset(raw, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert dataset: %w", err)
		}
		datasetDataList = append(datasetDataList, data)
	}

	// Save to database
	if err := s.saveDatasets(ctx, datasetDataList); err != nil {
		return nil, fmt.Errorf("failed to save datasets: %w", err)
	}

	// Collect dataset IDs for table ingestion
	datasetIDs := make([]string, 0, len(rawDatasets))
	for _, raw := range rawDatasets {
		datasetIDs = append(datasetIDs, raw.DatasetID)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		DatasetCount:   len(datasetDataList),
		DatasetIDs:     datasetIDs,
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveDatasets saves BigQuery datasets to the database with history tracking.
func (s *Service) saveDatasets(ctx context.Context, datasets []*DatasetData) error {
	if len(datasets) == 0 {
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

	for _, datasetData := range datasets {
		// Load existing dataset
		existing, err := tx.BronzeGCPBigQueryDataset.Query().
			Where(bronzegcpbigquerydataset.ID(datasetData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing dataset %s: %w", datasetData.ID, err)
		}

		// Compute diff
		diff := DiffDatasetData(existing, datasetData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPBigQueryDataset.UpdateOneID(datasetData.ID).
				SetCollectedAt(datasetData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for dataset %s: %w", datasetData.ID, err)
			}
			continue
		}

		// Create or update dataset
		if existing == nil {
			create := tx.BronzeGCPBigQueryDataset.Create().
				SetID(datasetData.ID).
				SetProjectID(datasetData.ProjectID).
				SetCollectedAt(datasetData.CollectedAt).
				SetFirstCollectedAt(datasetData.CollectedAt)

			if datasetData.FriendlyName != "" {
				create.SetFriendlyName(datasetData.FriendlyName)
			}
			if datasetData.Description != "" {
				create.SetDescription(datasetData.Description)
			}
			if datasetData.Location != "" {
				create.SetLocation(datasetData.Location)
			}
			if datasetData.DefaultTableExpirationMs != nil {
				create.SetDefaultTableExpirationMs(*datasetData.DefaultTableExpirationMs)
			}
			if datasetData.DefaultPartitionExpirationMs != nil {
				create.SetDefaultPartitionExpirationMs(*datasetData.DefaultPartitionExpirationMs)
			}
			if datasetData.LabelsJSON != nil {
				create.SetLabelsJSON(datasetData.LabelsJSON)
			}
			if datasetData.AccessJSON != nil {
				create.SetAccessJSON(datasetData.AccessJSON)
			}
			if datasetData.CreationTime != "" {
				create.SetCreationTime(datasetData.CreationTime)
			}
			if datasetData.LastModifiedTime != "" {
				create.SetLastModifiedTime(datasetData.LastModifiedTime)
			}
			if datasetData.Etag != "" {
				create.SetEtag(datasetData.Etag)
			}
			if datasetData.DefaultCollation != "" {
				create.SetDefaultCollation(datasetData.DefaultCollation)
			}
			if datasetData.MaxTimeTravelHours != nil {
				create.SetMaxTimeTravelHours(*datasetData.MaxTimeTravelHours)
			}
			if datasetData.DefaultEncryptionConfigurationJSON != nil {
				create.SetDefaultEncryptionConfigurationJSON(datasetData.DefaultEncryptionConfigurationJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create dataset %s: %w", datasetData.ID, err)
			}
		} else {
			update := tx.BronzeGCPBigQueryDataset.UpdateOneID(datasetData.ID).
				SetProjectID(datasetData.ProjectID).
				SetCollectedAt(datasetData.CollectedAt)

			if datasetData.FriendlyName != "" {
				update.SetFriendlyName(datasetData.FriendlyName)
			}
			if datasetData.Description != "" {
				update.SetDescription(datasetData.Description)
			}
			if datasetData.Location != "" {
				update.SetLocation(datasetData.Location)
			}
			if datasetData.DefaultTableExpirationMs != nil {
				update.SetDefaultTableExpirationMs(*datasetData.DefaultTableExpirationMs)
			}
			if datasetData.DefaultPartitionExpirationMs != nil {
				update.SetDefaultPartitionExpirationMs(*datasetData.DefaultPartitionExpirationMs)
			}
			if datasetData.LabelsJSON != nil {
				update.SetLabelsJSON(datasetData.LabelsJSON)
			}
			if datasetData.AccessJSON != nil {
				update.SetAccessJSON(datasetData.AccessJSON)
			}
			if datasetData.CreationTime != "" {
				update.SetCreationTime(datasetData.CreationTime)
			}
			if datasetData.LastModifiedTime != "" {
				update.SetLastModifiedTime(datasetData.LastModifiedTime)
			}
			if datasetData.Etag != "" {
				update.SetEtag(datasetData.Etag)
			}
			if datasetData.DefaultCollation != "" {
				update.SetDefaultCollation(datasetData.DefaultCollation)
			}
			if datasetData.MaxTimeTravelHours != nil {
				update.SetMaxTimeTravelHours(*datasetData.MaxTimeTravelHours)
			}
			if datasetData.DefaultEncryptionConfigurationJSON != nil {
				update.SetDefaultEncryptionConfigurationJSON(datasetData.DefaultEncryptionConfigurationJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update dataset %s: %w", datasetData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, datasetData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for dataset %s: %w", datasetData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, datasetData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for dataset %s: %w", datasetData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleDatasets removes datasets that were not collected in the latest run.
func (s *Service) DeleteStaleDatasets(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	staleDatasets, err := tx.BronzeGCPBigQueryDataset.Query().
		Where(
			bronzegcpbigquerydataset.ProjectID(projectID),
			bronzegcpbigquerydataset.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, ds := range staleDatasets {
		if err := s.history.CloseHistory(ctx, tx, ds.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for dataset %s: %w", ds.ID, err)
		}

		if err := tx.BronzeGCPBigQueryDataset.DeleteOne(ds).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete dataset %s: %w", ds.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
