package droplet

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzedodroplet"
)

// Service handles DigitalOcean Droplet ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new Droplet ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of Droplet ingestion.
type IngestResult struct {
	DropletCount   int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches all Droplets from DigitalOcean and saves them.
func (s *Service) Ingest(ctx context.Context, heartbeat func()) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	apiDroplets, err := s.client.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list droplets: %w", err)
	}

	if heartbeat != nil {
		heartbeat()
	}

	var allDroplets []*DropletData
	for _, v := range apiDroplets {
		allDroplets = append(allDroplets, ConvertDroplet(v, collectedAt))
	}

	if err := s.saveDroplets(ctx, allDroplets); err != nil {
		return nil, fmt.Errorf("save droplets: %w", err)
	}

	return &IngestResult{
		DropletCount:   len(allDroplets),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveDroplets(ctx context.Context, droplets []*DropletData) error {
	if len(droplets) == 0 {
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

	for _, data := range droplets {
		existing, err := tx.BronzeDODroplet.Query().
			Where(bronzedodroplet.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing droplet %s: %w", data.ResourceID, err)
		}

		diff := DiffDropletData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeDODroplet.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for droplet %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			if _, err := tx.BronzeDODroplet.Create().
				SetID(data.ResourceID).
				SetName(data.Name).
				SetMemory(data.Memory).
				SetVcpus(data.Vcpus).
				SetDisk(data.Disk).
				SetRegion(data.Region).
				SetSizeSlug(data.SizeSlug).
				SetStatus(data.Status).
				SetLocked(data.Locked).
				SetVpcUUID(data.VpcUUID).
				SetAPICreatedAt(data.APICreatedAt).
				SetImageJSON(data.ImageJSON).
				SetSizeJSON(data.SizeJSON).
				SetNetworksJSON(data.NetworksJSON).
				SetKernelJSON(data.KernelJSON).
				SetTagsJSON(data.TagsJSON).
				SetFeaturesJSON(data.FeaturesJSON).
				SetVolumeIdsJSON(data.VolumeIdsJSON).
				SetBackupIdsJSON(data.BackupIdsJSON).
				SetSnapshotIdsJSON(data.SnapshotIdsJSON).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create droplet %s: %w", data.ResourceID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for droplet %s: %w", data.ResourceID, err)
			}
		} else {
			if _, err := tx.BronzeDODroplet.UpdateOneID(data.ResourceID).
				SetName(data.Name).
				SetMemory(data.Memory).
				SetVcpus(data.Vcpus).
				SetDisk(data.Disk).
				SetRegion(data.Region).
				SetSizeSlug(data.SizeSlug).
				SetStatus(data.Status).
				SetLocked(data.Locked).
				SetVpcUUID(data.VpcUUID).
				SetAPICreatedAt(data.APICreatedAt).
				SetImageJSON(data.ImageJSON).
				SetSizeJSON(data.SizeJSON).
				SetNetworksJSON(data.NetworksJSON).
				SetKernelJSON(data.KernelJSON).
				SetTagsJSON(data.TagsJSON).
				SetFeaturesJSON(data.FeaturesJSON).
				SetVolumeIdsJSON(data.VolumeIdsJSON).
				SetBackupIdsJSON(data.BackupIdsJSON).
				SetSnapshotIdsJSON(data.SnapshotIdsJSON).
				SetCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update droplet %s: %w", data.ResourceID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for droplet %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStale removes Droplets that were not collected in the latest run.
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

	stale, err := tx.BronzeDODroplet.Query().
		Where(bronzedodroplet.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, doDroplet := range stale {
		if err := s.history.CloseHistory(ctx, tx, doDroplet.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for droplet %s: %w", doDroplet.ID, err)
		}

		if err := tx.BronzeDODroplet.DeleteOne(doDroplet).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete droplet %s: %w", doDroplet.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
