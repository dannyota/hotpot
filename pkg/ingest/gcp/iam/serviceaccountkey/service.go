package serviceaccountkey

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzegcpiamserviceaccountkey"
)

// Service handles GCP IAM service account key ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new service account key ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for service account key ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of service account key ingestion.
type IngestResult struct {
	ProjectID              string
	ServiceAccountKeyCount int
	CollectedAt            time.Time
	DurationMillis         int64
}

// Ingest fetches service account keys from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch service account keys from GCP
	keysWithAccounts, err := s.client.ListServiceAccountKeys(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list service account keys: %w", err)
	}

	// Convert to data structs
	keyDataList := make([]*ServiceAccountKeyData, 0, len(keysWithAccounts))
	for _, kwa := range keysWithAccounts {
		data := ConvertServiceAccountKey(kwa, params.ProjectID, collectedAt)
		keyDataList = append(keyDataList, data)
	}

	// Save to database
	if err := s.saveServiceAccountKeys(ctx, keyDataList); err != nil {
		return nil, fmt.Errorf("failed to save service account keys: %w", err)
	}

	return &IngestResult{
		ProjectID:              params.ProjectID,
		ServiceAccountKeyCount: len(keyDataList),
		CollectedAt:            collectedAt,
		DurationMillis:         time.Since(startTime).Milliseconds(),
	}, nil
}

// saveServiceAccountKeys saves service account keys to the database with history tracking.
func (s *Service) saveServiceAccountKeys(ctx context.Context, keys []*ServiceAccountKeyData) error {
	if len(keys) == 0 {
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

	for _, keyData := range keys {
		// Load existing service account key
		existing, err := tx.BronzeGCPIAMServiceAccountKey.Query().
			Where(bronzegcpiamserviceaccountkey.ID(keyData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing key %s: %w", keyData.ID, err)
		}

		// Compute diff
		diff := DiffServiceAccountKeyData(existing, keyData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPIAMServiceAccountKey.UpdateOneID(keyData.ID).
				SetCollectedAt(keyData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for key %s: %w", keyData.ID, err)
			}
			continue
		}

		// Create or update service account key
		if existing == nil {
			// Create new service account key
			create := tx.BronzeGCPIAMServiceAccountKey.Create().
				SetID(keyData.ID).
				SetName(keyData.Name).
				SetServiceAccountEmail(keyData.ServiceAccountEmail).
				SetDisabled(keyData.Disabled).
				SetProjectID(keyData.ProjectID).
				SetCollectedAt(keyData.CollectedAt).
				SetFirstCollectedAt(keyData.CollectedAt)

			if keyData.KeyOrigin != "" {
				create.SetKeyOrigin(keyData.KeyOrigin)
			}
			if keyData.KeyType != "" {
				create.SetKeyType(keyData.KeyType)
			}
			if keyData.KeyAlgorithm != "" {
				create.SetKeyAlgorithm(keyData.KeyAlgorithm)
			}
			if !keyData.ValidAfterTime.IsZero() {
				create.SetValidAfterTime(keyData.ValidAfterTime)
			}
			if !keyData.ValidBeforeTime.IsZero() {
				create.SetValidBeforeTime(keyData.ValidBeforeTime)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create service account key %s: %w", keyData.ID, err)
			}
		} else {
			// Update existing service account key
			update := tx.BronzeGCPIAMServiceAccountKey.UpdateOneID(keyData.ID).
				SetName(keyData.Name).
				SetServiceAccountEmail(keyData.ServiceAccountEmail).
				SetDisabled(keyData.Disabled).
				SetProjectID(keyData.ProjectID).
				SetCollectedAt(keyData.CollectedAt)

			if keyData.KeyOrigin != "" {
				update.SetKeyOrigin(keyData.KeyOrigin)
			}
			if keyData.KeyType != "" {
				update.SetKeyType(keyData.KeyType)
			}
			if keyData.KeyAlgorithm != "" {
				update.SetKeyAlgorithm(keyData.KeyAlgorithm)
			}
			if !keyData.ValidAfterTime.IsZero() {
				update.SetValidAfterTime(keyData.ValidAfterTime)
			}
			if !keyData.ValidBeforeTime.IsZero() {
				update.SetValidBeforeTime(keyData.ValidBeforeTime)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update service account key %s: %w", keyData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, keyData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for key %s: %w", keyData.ID, err)
			}
		} else if diff.IsChanged {
			if err := s.history.UpdateHistory(ctx, tx, existing, keyData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for key %s: %w", keyData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleKeys removes service account keys that were not collected in the latest run.
// Also closes history records for deleted keys.
func (s *Service) DeleteStaleKeys(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale keys
	staleKeys, err := tx.BronzeGCPIAMServiceAccountKey.Query().
		Where(
			bronzegcpiamserviceaccountkey.ProjectID(projectID),
			bronzegcpiamserviceaccountkey.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale key
	for _, key := range staleKeys {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, key.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for key %s: %w", key.ID, err)
		}

		// Delete key
		if err := tx.BronzeGCPIAMServiceAccountKey.DeleteOne(key).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete key %s: %w", key.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
