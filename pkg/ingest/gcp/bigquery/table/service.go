package table

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpbigquerytable"
)

// Service handles BigQuery table ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new BigQuery table ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for table ingestion.
type IngestParams struct {
	ProjectID  string
	DatasetIDs []string
}

// IngestResult contains the result of table ingestion.
type IngestResult struct {
	ProjectID      string
	TableCount     int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches BigQuery tables from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch tables from GCP
	rawTables, err := s.client.ListTables(ctx, params.DatasetIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}

	// Convert to table data
	tableDataList := make([]*TableData, 0, len(rawTables))
	for _, raw := range rawTables {
		data, err := ConvertTable(raw, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert table: %w", err)
		}
		tableDataList = append(tableDataList, data)
	}

	// Save to database
	if err := s.saveTables(ctx, tableDataList); err != nil {
		return nil, fmt.Errorf("failed to save tables: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		TableCount:     len(tableDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveTables saves BigQuery tables to the database with history tracking.
func (s *Service) saveTables(ctx context.Context, tables []*TableData) error {
	if len(tables) == 0 {
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

	for _, tableData := range tables {
		// Load existing table
		existing, err := tx.BronzeGCPBigQueryTable.Query().
			Where(bronzegcpbigquerytable.ID(tableData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing table %s: %w", tableData.ID, err)
		}

		// Compute diff
		diff := DiffTableData(existing, tableData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPBigQueryTable.UpdateOneID(tableData.ID).
				SetCollectedAt(tableData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for table %s: %w", tableData.ID, err)
			}
			continue
		}

		// Create or update table
		if existing == nil {
			create := tx.BronzeGCPBigQueryTable.Create().
				SetID(tableData.ID).
				SetDatasetID(tableData.DatasetID).
				SetProjectID(tableData.ProjectID).
				SetCollectedAt(tableData.CollectedAt).
				SetFirstCollectedAt(tableData.CollectedAt).
				SetRequirePartitionFilter(tableData.RequirePartitionFilter)

			if tableData.FriendlyName != "" {
				create.SetFriendlyName(tableData.FriendlyName)
			}
			if tableData.Description != "" {
				create.SetDescription(tableData.Description)
			}
			if tableData.SchemaJSON != nil {
				create.SetSchemaJSON(tableData.SchemaJSON)
			}
			if tableData.NumBytes != nil {
				create.SetNumBytes(*tableData.NumBytes)
			}
			if tableData.NumLongTermBytes != nil {
				create.SetNumLongTermBytes(*tableData.NumLongTermBytes)
			}
			if tableData.NumRows != nil {
				create.SetNumRows(*tableData.NumRows)
			}
			if tableData.CreationTime != "" {
				create.SetCreationTime(tableData.CreationTime)
			}
			if tableData.ExpirationTime != "" {
				create.SetExpirationTime(tableData.ExpirationTime)
			}
			if tableData.LastModifiedTime != "" {
				create.SetLastModifiedTime(tableData.LastModifiedTime)
			}
			if tableData.TableType != "" {
				create.SetTableType(tableData.TableType)
			}
			if tableData.LabelsJSON != nil {
				create.SetLabelsJSON(tableData.LabelsJSON)
			}
			if tableData.EncryptionConfigurationJSON != nil {
				create.SetEncryptionConfigurationJSON(tableData.EncryptionConfigurationJSON)
			}
			if tableData.TimePartitioningJSON != nil {
				create.SetTimePartitioningJSON(tableData.TimePartitioningJSON)
			}
			if tableData.RangePartitioningJSON != nil {
				create.SetRangePartitioningJSON(tableData.RangePartitioningJSON)
			}
			if tableData.ClusteringJSON != nil {
				create.SetClusteringJSON(tableData.ClusteringJSON)
			}
			if tableData.Etag != "" {
				create.SetEtag(tableData.Etag)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create table %s: %w", tableData.ID, err)
			}
		} else {
			update := tx.BronzeGCPBigQueryTable.UpdateOneID(tableData.ID).
				SetDatasetID(tableData.DatasetID).
				SetProjectID(tableData.ProjectID).
				SetCollectedAt(tableData.CollectedAt).
				SetRequirePartitionFilter(tableData.RequirePartitionFilter)

			if tableData.FriendlyName != "" {
				update.SetFriendlyName(tableData.FriendlyName)
			}
			if tableData.Description != "" {
				update.SetDescription(tableData.Description)
			}
			if tableData.SchemaJSON != nil {
				update.SetSchemaJSON(tableData.SchemaJSON)
			}
			if tableData.NumBytes != nil {
				update.SetNumBytes(*tableData.NumBytes)
			}
			if tableData.NumLongTermBytes != nil {
				update.SetNumLongTermBytes(*tableData.NumLongTermBytes)
			}
			if tableData.NumRows != nil {
				update.SetNumRows(*tableData.NumRows)
			}
			if tableData.CreationTime != "" {
				update.SetCreationTime(tableData.CreationTime)
			}
			if tableData.ExpirationTime != "" {
				update.SetExpirationTime(tableData.ExpirationTime)
			}
			if tableData.LastModifiedTime != "" {
				update.SetLastModifiedTime(tableData.LastModifiedTime)
			}
			if tableData.TableType != "" {
				update.SetTableType(tableData.TableType)
			}
			if tableData.LabelsJSON != nil {
				update.SetLabelsJSON(tableData.LabelsJSON)
			}
			if tableData.EncryptionConfigurationJSON != nil {
				update.SetEncryptionConfigurationJSON(tableData.EncryptionConfigurationJSON)
			}
			if tableData.TimePartitioningJSON != nil {
				update.SetTimePartitioningJSON(tableData.TimePartitioningJSON)
			}
			if tableData.RangePartitioningJSON != nil {
				update.SetRangePartitioningJSON(tableData.RangePartitioningJSON)
			}
			if tableData.ClusteringJSON != nil {
				update.SetClusteringJSON(tableData.ClusteringJSON)
			}
			if tableData.Etag != "" {
				update.SetEtag(tableData.Etag)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update table %s: %w", tableData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, tableData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for table %s: %w", tableData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, tableData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for table %s: %w", tableData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleTables removes tables that were not collected in the latest run.
func (s *Service) DeleteStaleTables(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	staleTables, err := tx.BronzeGCPBigQueryTable.Query().
		Where(
			bronzegcpbigquerytable.ProjectID(projectID),
			bronzegcpbigquerytable.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, tbl := range staleTables {
		if err := s.history.CloseHistory(ctx, tx, tbl.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for table %s: %w", tbl.ID, err)
		}

		if err := tx.BronzeGCPBigQueryTable.DeleteOne(tbl).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete table %s: %w", tbl.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
