package cryptokey

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpkmscryptokey"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpkmskeyring"
)

// Service handles GCP KMS crypto key ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new crypto key ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for crypto key ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of crypto key ingestion.
type IngestResult struct {
	ProjectID      string
	CryptoKeyCount int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches crypto keys from GCP and stores them in the bronze layer.
// Queries key rings from the database and lists crypto keys for each.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Query key rings from database
	keyRings, err := s.entClient.BronzeGCPKMSKeyRing.Query().
		Where(bronzegcpkmskeyring.ProjectID(params.ProjectID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query key rings: %w", err)
	}

	var allData []*CryptoKeyData
	for _, kr := range keyRings {
		keys, err := s.client.ListCryptoKeys(ctx, kr.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to list crypto keys for key ring %s: %w", kr.ID, err)
		}

		for _, key := range keys {
			data, err := ConvertCryptoKey(key, params.ProjectID, collectedAt)
			if err != nil {
				return nil, fmt.Errorf("failed to convert crypto key: %w", err)
			}
			allData = append(allData, data)
		}
	}

	if err := s.saveCryptoKeys(ctx, allData); err != nil {
		return nil, fmt.Errorf("failed to save crypto keys: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		CryptoKeyCount: len(allData),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveCryptoKeys(ctx context.Context, keys []*CryptoKeyData) error {
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

	for _, data := range keys {
		existing, err := tx.BronzeGCPKMSCryptoKey.Query().
			Where(bronzegcpkmscryptokey.ID(data.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing crypto key %s: %w", data.Name, err)
		}

		diff := DiffCryptoKeyData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPKMSCryptoKey.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for crypto key %s: %w", data.Name, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeGCPKMSCryptoKey.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetPurpose(data.Purpose).
				SetCreateTime(data.CreateTime).
				SetNextRotationTime(data.NextRotationTime).
				SetRotationPeriod(data.RotationPeriod).
				SetDestroyScheduledDuration(data.DestroyScheduledDuration).
				SetImportOnly(data.ImportOnly).
				SetCryptoKeyBackend(data.CryptoKeyBackend).
				SetProjectID(data.ProjectID).
				SetLocation(data.Location).
				SetKeyRingName(data.KeyRingName).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.VersionTemplateJSON != nil {
				create.SetVersionTemplateJSON(data.VersionTemplateJSON)
			}
			if data.PrimaryJSON != nil {
				create.SetPrimaryJSON(data.PrimaryJSON)
			}
			if data.LabelsJSON != nil {
				create.SetLabelsJSON(data.LabelsJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create crypto key %s: %w", data.Name, err)
			}
		} else {
			update := tx.BronzeGCPKMSCryptoKey.UpdateOneID(data.ID).
				SetName(data.Name).
				SetPurpose(data.Purpose).
				SetCreateTime(data.CreateTime).
				SetNextRotationTime(data.NextRotationTime).
				SetRotationPeriod(data.RotationPeriod).
				SetDestroyScheduledDuration(data.DestroyScheduledDuration).
				SetImportOnly(data.ImportOnly).
				SetCryptoKeyBackend(data.CryptoKeyBackend).
				SetProjectID(data.ProjectID).
				SetLocation(data.Location).
				SetKeyRingName(data.KeyRingName).
				SetCollectedAt(data.CollectedAt)

			if data.VersionTemplateJSON != nil {
				update.SetVersionTemplateJSON(data.VersionTemplateJSON)
			}
			if data.PrimaryJSON != nil {
				update.SetPrimaryJSON(data.PrimaryJSON)
			}
			if data.LabelsJSON != nil {
				update.SetLabelsJSON(data.LabelsJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update crypto key %s: %w", data.Name, err)
			}
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for crypto key %s: %w", data.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for crypto key %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleCryptoKeys removes crypto keys that were not collected in the latest run.
func (s *Service) DeleteStaleCryptoKeys(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGCPKMSCryptoKey.Query().
		Where(
			bronzegcpkmscryptokey.ProjectID(projectID),
			bronzegcpkmscryptokey.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, k := range stale {
		if err := s.history.CloseHistory(ctx, tx, k.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for crypto key %s: %w", k.ID, err)
		}

		if err := tx.BronzeGCPKMSCryptoKey.DeleteOne(k).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete crypto key %s: %w", k.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
