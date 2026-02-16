package keyring

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpkmskeyring"
)

// Service handles GCP KMS key ring ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new key ring ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for key ring ingestion.
type IngestParams struct {
	ProjectID string
	Locations []string
}

// IngestResult contains the result of key ring ingestion.
type IngestResult struct {
	ProjectID      string
	KeyRingCount   int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches key rings from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	keyRings, err := s.client.ListKeyRings(ctx, params.ProjectID, params.Locations)
	if err != nil {
		return nil, fmt.Errorf("failed to list key rings: %w", err)
	}

	dataList := make([]*KeyRingData, 0, len(keyRings))
	for _, kr := range keyRings {
		data := ConvertKeyRing(kr.GetName(), kr.GetCreateTime().AsTime().Format(time.RFC3339), params.ProjectID, collectedAt)
		dataList = append(dataList, data)
	}

	if err := s.saveKeyRings(ctx, dataList); err != nil {
		return nil, fmt.Errorf("failed to save key rings: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		KeyRingCount:   len(dataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveKeyRings(ctx context.Context, keyRings []*KeyRingData) error {
	if len(keyRings) == 0 {
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

	for _, data := range keyRings {
		existing, err := tx.BronzeGCPKMSKeyRing.Query().
			Where(bronzegcpkmskeyring.ID(data.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing key ring %s: %w", data.Name, err)
		}

		diff := DiffKeyRingData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPKMSKeyRing.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for key ring %s: %w", data.Name, err)
			}
			continue
		}

		if existing == nil {
			_, err = tx.BronzeGCPKMSKeyRing.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetCreateTime(data.CreateTime).
				SetProjectID(data.ProjectID).
				SetLocation(data.Location).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create key ring %s: %w", data.Name, err)
			}
		} else {
			_, err = tx.BronzeGCPKMSKeyRing.UpdateOneID(data.ID).
				SetName(data.Name).
				SetCreateTime(data.CreateTime).
				SetProjectID(data.ProjectID).
				SetLocation(data.Location).
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update key ring %s: %w", data.Name, err)
			}
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for key ring %s: %w", data.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for key ring %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleKeyRings removes key rings that were not collected in the latest run.
func (s *Service) DeleteStaleKeyRings(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGCPKMSKeyRing.Query().
		Where(
			bronzegcpkmskeyring.ProjectID(projectID),
			bronzegcpkmskeyring.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, kr := range stale {
		if err := s.history.CloseHistory(ctx, tx, kr.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for key ring %s: %w", kr.ID, err)
		}

		if err := tx.BronzeGCPKMSKeyRing.DeleteOne(kr).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete key ring %s: %w", kr.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
