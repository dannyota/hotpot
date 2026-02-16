package database

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpspannerdatabase"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpspannerinstance"
)

// Service handles Spanner database ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new Spanner database ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for database ingestion.
type IngestParams struct {
	ProjectID     string
	InstanceNames []string
}

// IngestResult contains the result of database ingestion.
type IngestResult struct {
	ProjectID      string
	DatabaseCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches Spanner databases from GCP and stores them in the bronze layer.
// Databases are listed per-instance using the provided instance names.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch databases for each instance
	var databaseDataList []*DatabaseData
	for _, instanceName := range params.InstanceNames {
		databases, err := s.client.ListDatabases(ctx, instanceName)
		if err != nil {
			// Skip individual instance failures
			continue
		}

		for _, db := range databases {
			data, err := ConvertDatabase(db, instanceName, params.ProjectID, collectedAt)
			if err != nil {
				return nil, fmt.Errorf("failed to convert Spanner database: %w", err)
			}
			databaseDataList = append(databaseDataList, data)
		}
	}

	// Save to database
	if err := s.saveDatabases(ctx, databaseDataList); err != nil {
		return nil, fmt.Errorf("failed to save Spanner databases: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		DatabaseCount:  len(databaseDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveDatabases saves Spanner databases to the database with history tracking.
func (s *Service) saveDatabases(ctx context.Context, databases []*DatabaseData) error {
	if len(databases) == 0 {
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

	for _, dbData := range databases {
		// Load existing database
		existing, err := tx.BronzeGCPSpannerDatabase.Query().
			Where(bronzegcpspannerdatabase.ID(dbData.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing Spanner database %s: %w", dbData.ResourceID, err)
		}

		// Compute diff
		diff := DiffDatabaseData(existing, dbData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPSpannerDatabase.UpdateOneID(dbData.ResourceID).
				SetCollectedAt(dbData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for Spanner database %s: %w", dbData.ResourceID, err)
			}
			continue
		}

		// Look up parent instance for edge
		parentInstance, err := tx.BronzeGCPSpannerInstance.Query().
			Where(bronzegcpspannerinstance.ID(dbData.InstanceName)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to find parent instance %s: %w", dbData.InstanceName, err)
		}

		// Create or update database
		if existing == nil {
			create := tx.BronzeGCPSpannerDatabase.Create().
				SetID(dbData.ResourceID).
				SetName(dbData.Name).
				SetState(dbData.State).
				SetDatabaseDialect(dbData.DatabaseDialect).
				SetEnableDropProtection(dbData.EnableDropProtection).
				SetReconciling(dbData.Reconciling).
				SetProjectID(dbData.ProjectID).
				SetCollectedAt(dbData.CollectedAt).
				SetFirstCollectedAt(dbData.CollectedAt)

			if dbData.CreateTime != "" {
				create.SetCreateTime(dbData.CreateTime)
			}
			if dbData.VersionRetentionPeriod != "" {
				create.SetVersionRetentionPeriod(dbData.VersionRetentionPeriod)
			}
			if dbData.EarliestVersionTime != "" {
				create.SetEarliestVersionTime(dbData.EarliestVersionTime)
			}
			if dbData.DefaultLeader != "" {
				create.SetDefaultLeader(dbData.DefaultLeader)
			}
			if dbData.InstanceName != "" {
				create.SetInstanceName(dbData.InstanceName)
			}
			if dbData.RestoreInfoJSON != nil {
				create.SetRestoreInfoJSON(dbData.RestoreInfoJSON)
			}
			if dbData.EncryptionConfigJSON != nil {
				create.SetEncryptionConfigJSON(dbData.EncryptionConfigJSON)
			}
			if dbData.EncryptionInfoJSON != nil {
				create.SetEncryptionInfoJSON(dbData.EncryptionInfoJSON)
			}
			if parentInstance != nil {
				create.SetInstance(parentInstance)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create Spanner database %s: %w", dbData.ResourceID, err)
			}
		} else {
			update := tx.BronzeGCPSpannerDatabase.UpdateOneID(dbData.ResourceID).
				SetName(dbData.Name).
				SetState(dbData.State).
				SetDatabaseDialect(dbData.DatabaseDialect).
				SetEnableDropProtection(dbData.EnableDropProtection).
				SetReconciling(dbData.Reconciling).
				SetProjectID(dbData.ProjectID).
				SetCollectedAt(dbData.CollectedAt)

			if dbData.CreateTime != "" {
				update.SetCreateTime(dbData.CreateTime)
			}
			if dbData.VersionRetentionPeriod != "" {
				update.SetVersionRetentionPeriod(dbData.VersionRetentionPeriod)
			}
			if dbData.EarliestVersionTime != "" {
				update.SetEarliestVersionTime(dbData.EarliestVersionTime)
			}
			if dbData.DefaultLeader != "" {
				update.SetDefaultLeader(dbData.DefaultLeader)
			}
			if dbData.InstanceName != "" {
				update.SetInstanceName(dbData.InstanceName)
			}
			if dbData.RestoreInfoJSON != nil {
				update.SetRestoreInfoJSON(dbData.RestoreInfoJSON)
			}
			if dbData.EncryptionConfigJSON != nil {
				update.SetEncryptionConfigJSON(dbData.EncryptionConfigJSON)
			}
			if dbData.EncryptionInfoJSON != nil {
				update.SetEncryptionInfoJSON(dbData.EncryptionInfoJSON)
			}
			if parentInstance != nil {
				update.SetInstance(parentInstance)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update Spanner database %s: %w", dbData.ResourceID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, dbData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for Spanner database %s: %w", dbData.ResourceID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, dbData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for Spanner database %s: %w", dbData.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleDatabases removes databases that were not collected in the latest run.
func (s *Service) DeleteStaleDatabases(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	staleDatabases, err := tx.BronzeGCPSpannerDatabase.Query().
		Where(
			bronzegcpspannerdatabase.ProjectID(projectID),
			bronzegcpspannerdatabase.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, db := range staleDatabases {
		if err := s.history.CloseHistory(ctx, tx, db.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for Spanner database %s: %w", db.ID, err)
		}

		if err := tx.BronzeGCPSpannerDatabase.DeleteOne(db).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete Spanner database %s: %w", db.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
