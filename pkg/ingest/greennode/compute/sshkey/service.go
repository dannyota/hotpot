package sshkey

import (
	"context"
	"fmt"
	"time"

	entcompute "danny.vn/hotpot/pkg/storage/ent/greennode/compute"
	"danny.vn/hotpot/pkg/storage/ent/greennode/compute/bronzegreennodecomputesshkey"
)

// Service handles GreenNode SSH key ingestion.
type Service struct {
	client    *Client
	entClient *entcompute.Client
	history   *HistoryService
}

// NewService creates a new SSH key ingestion service.
func NewService(client *Client, entClient *entcompute.Client) *Service {
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

// Ingest fetches SSH keys from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID, region string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	keys, err := s.client.ListSSHKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("list ssh keys: %w", err)
	}

	keyDataList := make([]*SSHKeyData, 0, len(keys))
	for _, k := range keys {
		keyDataList = append(keyDataList, ConvertSSHKey(k, projectID, region, collectedAt))
	}

	if err := s.saveSSHKeys(ctx, keyDataList); err != nil {
		return nil, fmt.Errorf("save ssh keys: %w", err)
	}

	return &IngestResult{
		KeyCount:       len(keyDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveSSHKeys(ctx context.Context, keys []*SSHKeyData) error {
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
		existing, err := tx.BronzeGreenNodeComputeSSHKey.Query().
			Where(bronzegreennodecomputesshkey.ID(data.ID)).
			First(ctx)
		if err != nil && !entcompute.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing ssh key %s: %w", data.Name, err)
		}

		diff := DiffSSHKeyData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodeComputeSSHKey.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for ssh key %s: %w", data.Name, err)
			}
			continue
		}

		if existing == nil {
			_, err = tx.BronzeGreenNodeComputeSSHKey.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetCreatedAtAPI(data.CreatedAtAPI).
				SetPubKey(data.PubKey).
				SetStatus(data.Status).
				SetRegion(data.Region).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create ssh key %s: %w", data.Name, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for ssh key %s: %w", data.Name, err)
			}
		} else {
			_, err = tx.BronzeGreenNodeComputeSSHKey.UpdateOneID(data.ID).
				SetName(data.Name).
				SetCreatedAtAPI(data.CreatedAtAPI).
				SetPubKey(data.PubKey).
				SetStatus(data.Status).
				SetRegion(data.Region).
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update ssh key %s: %w", data.Name, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for ssh key %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// DeleteStaleSSHKeys removes SSH keys not collected in the latest run for the given region.
func (s *Service) DeleteStaleSSHKeys(ctx context.Context, projectID, region string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGreenNodeComputeSSHKey.Query().
		Where(
			bronzegreennodecomputesshkey.ProjectID(projectID),
			bronzegreennodecomputesshkey.Region(region),
			bronzegreennodecomputesshkey.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale ssh keys: %w", err)
	}

	for _, k := range stale {
		if err := s.history.CloseHistory(ctx, tx, k.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for ssh key %s: %w", k.ID, err)
		}
		if err := tx.BronzeGreenNodeComputeSSHKey.DeleteOneID(k.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete ssh key %s: %w", k.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
