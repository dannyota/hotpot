package volume

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzedovolume"
)

// Service handles DigitalOcean Volume ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new Volume ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of Volume ingestion.
type IngestResult struct {
	VolumeCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches all Volumes from DigitalOcean and saves them.
func (s *Service) Ingest(ctx context.Context, heartbeat func()) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	apiVolumes, err := s.client.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list volumes: %w", err)
	}

	if heartbeat != nil {
		heartbeat()
	}

	var allVolumes []*VolumeData
	for _, v := range apiVolumes {
		allVolumes = append(allVolumes, ConvertVolume(v, collectedAt))
	}

	if err := s.saveVolumes(ctx, allVolumes); err != nil {
		return nil, fmt.Errorf("save volumes: %w", err)
	}

	return &IngestResult{
		VolumeCount:    len(allVolumes),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveVolumes(ctx context.Context, volumes []*VolumeData) error {
	if len(volumes) == 0 {
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

	for _, data := range volumes {
		existing, err := tx.BronzeDOVolume.Query().
			Where(bronzedovolume.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing volume %s: %w", data.ResourceID, err)
		}

		diff := DiffVolumeData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeDOVolume.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for volume %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeDOVolume.Create().
				SetID(data.ResourceID).
				SetName(data.Name).
				SetRegion(data.Region).
				SetSizeGigabytes(data.SizeGigabytes).
				SetDescription(data.Description).
				SetDropletIdsJSON(data.DropletIdsJSON).
				SetFilesystemType(data.FilesystemType).
				SetFilesystemLabel(data.FilesystemLabel).
				SetTagsJSON(data.TagsJSON).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.APICreatedAt != nil {
				create.SetAPICreatedAt(*data.APICreatedAt)
			}

			if _, err := create.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create volume %s: %w", data.ResourceID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for volume %s: %w", data.ResourceID, err)
			}
		} else {
			update := tx.BronzeDOVolume.UpdateOneID(data.ResourceID).
				SetName(data.Name).
				SetRegion(data.Region).
				SetSizeGigabytes(data.SizeGigabytes).
				SetDescription(data.Description).
				SetDropletIdsJSON(data.DropletIdsJSON).
				SetFilesystemType(data.FilesystemType).
				SetFilesystemLabel(data.FilesystemLabel).
				SetTagsJSON(data.TagsJSON).
				SetCollectedAt(data.CollectedAt)

			if data.APICreatedAt != nil {
				update.SetAPICreatedAt(*data.APICreatedAt)
			}

			if _, err := update.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update volume %s: %w", data.ResourceID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for volume %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStale removes Volumes that were not collected in the latest run.
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

	stale, err := tx.BronzeDOVolume.Query().
		Where(bronzedovolume.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, doVolume := range stale {
		if err := s.history.CloseHistory(ctx, tx, doVolume.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for volume %s: %w", doVolume.ID, err)
		}

		if err := tx.BronzeDOVolume.DeleteOne(doVolume).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete volume %s: %w", doVolume.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
