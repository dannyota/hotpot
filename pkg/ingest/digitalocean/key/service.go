package key

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzedokey"
)

// Service handles DigitalOcean SSH key ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new SSH key ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of SSH key ingestion.
type IngestResult struct {
	KeyCount       int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches all SSH keys from DigitalOcean and saves them.
func (s *Service) Ingest(ctx context.Context, heartbeat func()) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	apiKeys, err := s.client.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list keys: %w", err)
	}

	if heartbeat != nil {
		heartbeat()
	}

	var allKeys []*KeyData
	for _, v := range apiKeys {
		allKeys = append(allKeys, ConvertKey(v, collectedAt))
	}

	if err := s.saveKeys(ctx, allKeys); err != nil {
		return nil, fmt.Errorf("save keys: %w", err)
	}

	return &IngestResult{
		KeyCount:       len(allKeys),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveKeys(ctx context.Context, keys []*KeyData) error {
	if len(keys) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, data := range keys {
		existing, err := tx.BronzeDOKey.Query().
			Where(bronzedokey.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing key %s: %w", data.ResourceID, err)
		}

		diff := DiffKeyData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeDOKey.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for key %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			if _, err := tx.BronzeDOKey.Create().
				SetID(data.ResourceID).
				SetName(data.Name).
				SetFingerprint(data.Fingerprint).
				SetPublicKey(data.PublicKey).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create key %s: %w", data.ResourceID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for key %s: %w", data.ResourceID, err)
			}
		} else {
			if _, err := tx.BronzeDOKey.UpdateOneID(data.ResourceID).
				SetName(data.Name).
				SetFingerprint(data.Fingerprint).
				SetPublicKey(data.PublicKey).
				SetCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update key %s: %w", data.ResourceID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for key %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStale removes SSH keys that were not collected in the latest run.
func (s *Service) DeleteStale(ctx context.Context, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	stale, err := tx.BronzeDOKey.Query().
		Where(bronzedokey.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, doKey := range stale {
		if err := s.history.CloseHistory(ctx, tx, doKey.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for key %s: %w", doKey.ID, err)
		}

		if err := tx.BronzeDOKey.DeleteOne(doKey).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete key %s: %w", doKey.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
