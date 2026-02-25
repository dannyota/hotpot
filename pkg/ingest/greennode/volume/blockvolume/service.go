package blockvolume

import (
	"context"
	"fmt"
	"time"

	entvol "github.com/dannyota/hotpot/pkg/storage/ent/greennode/volume"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/volume/bronzegreennodevolumeblockvolume"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/volume/bronzegreennodevolumesnapshot"
)

// Service handles GreenNode block volume ingestion.
type Service struct {
	client    *Client
	entClient *entvol.Client
	history   *HistoryService
}

// NewService creates a new block volume ingestion service.
func NewService(client *Client, entClient *entvol.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of block volume ingestion.
type IngestResult struct {
	BlockVolumeCount int
	CollectedAt      time.Time
	DurationMillis   int64
}

// Ingest fetches block volumes from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID, region string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	volumes, err := s.client.ListBlockVolumes(ctx)
	if err != nil {
		return nil, fmt.Errorf("list block volumes: %w", err)
	}

	dataList := make([]*BlockVolumeData, 0, len(volumes))
	for _, v := range volumes {
		data, err := ConvertBlockVolume(v, projectID, region, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("convert block volume: %w", err)
		}

		// Fetch snapshots for this volume
		snapshots, err := s.client.ListSnapshotsByBlockVolumeID(ctx, v.ID)
		if err != nil {
			return nil, fmt.Errorf("list snapshots for volume %s: %w", v.ID, err)
		}
		data.Snapshots = ConvertSnapshots(snapshots)

		dataList = append(dataList, data)
	}

	if err := s.saveBlockVolumes(ctx, dataList); err != nil {
		return nil, fmt.Errorf("save block volumes: %w", err)
	}

	return &IngestResult{
		BlockVolumeCount: len(dataList),
		CollectedAt:      collectedAt,
		DurationMillis:   time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveBlockVolumes(ctx context.Context, volumes []*BlockVolumeData) error {
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
		existing, err := tx.BronzeGreenNodeVolumeBlockVolume.Query().
			Where(bronzegreennodevolumeblockvolume.ID(data.ID)).
			WithSnapshots().
			First(ctx)
		if err != nil && !entvol.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing block volume %s: %w", data.Name, err)
		}

		diff := DiffBlockVolumeData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodeVolumeBlockVolume.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for block volume %s: %w", data.Name, err)
			}
			continue
		}

		if existing != nil {
			if err := s.deleteBlockVolumeChildren(ctx, tx, data.ID); err != nil {
				tx.Rollback()
				return fmt.Errorf("delete children for block volume %s: %w", data.Name, err)
			}
		}

		var savedVolume *entvol.BronzeGreenNodeVolumeBlockVolume
		if existing == nil {
			create := tx.BronzeGreenNodeVolumeBlockVolume.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetVolumeTypeID(data.VolumeTypeID).
				SetClusterID(data.ClusterID).
				SetVMID(data.VMID).
				SetSize(data.Size).
				SetIopsID(data.IopsID).
				SetStatus(data.Status).
				SetCreatedAtAPI(data.CreatedAtAPI).
				SetUpdatedAtAPI(data.UpdatedAtAPI).
				SetPersistentVolume(data.PersistentVolume).
				SetUnderID(data.UnderID).
				SetMigrateState(data.MigrateState).
				SetMultiAttach(data.MultiAttach).
				SetZoneID(data.ZoneID).
				SetRegion(data.Region).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.AttachedMachineJSON != nil {
				create.SetAttachedMachineJSON(data.AttachedMachineJSON)
			}

			savedVolume, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create block volume %s: %w", data.Name, err)
			}
		} else {
			update := tx.BronzeGreenNodeVolumeBlockVolume.UpdateOneID(data.ID).
				SetName(data.Name).
				SetVolumeTypeID(data.VolumeTypeID).
				SetClusterID(data.ClusterID).
				SetVMID(data.VMID).
				SetSize(data.Size).
				SetIopsID(data.IopsID).
				SetStatus(data.Status).
				SetCreatedAtAPI(data.CreatedAtAPI).
				SetUpdatedAtAPI(data.UpdatedAtAPI).
				SetPersistentVolume(data.PersistentVolume).
				SetUnderID(data.UnderID).
				SetMigrateState(data.MigrateState).
				SetMultiAttach(data.MultiAttach).
				SetZoneID(data.ZoneID).
				SetRegion(data.Region).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt)

			if data.AttachedMachineJSON != nil {
				update.SetAttachedMachineJSON(data.AttachedMachineJSON)
			}

			savedVolume, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update block volume %s: %w", data.Name, err)
			}
		}

		if err := s.createBlockVolumeChildren(ctx, tx, savedVolume, data); err != nil {
			tx.Rollback()
			return fmt.Errorf("create children for block volume %s: %w", data.Name, err)
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for block volume %s: %w", data.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, data, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for block volume %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (s *Service) deleteBlockVolumeChildren(ctx context.Context, tx *entvol.Tx, volumeID string) error {
	_, err := tx.BronzeGreenNodeVolumeSnapshot.Delete().
		Where(bronzegreennodevolumesnapshot.HasBlockVolumeWith(bronzegreennodevolumeblockvolume.ID(volumeID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete snapshots: %w", err)
	}
	return nil
}

func (s *Service) createBlockVolumeChildren(ctx context.Context, tx *entvol.Tx, volume *entvol.BronzeGreenNodeVolumeBlockVolume, data *BlockVolumeData) error {
	for _, snap := range data.Snapshots {
		_, err := tx.BronzeGreenNodeVolumeSnapshot.Create().
			SetBlockVolume(volume).
			SetSnapshotID(snap.SnapshotID).
			SetName(snap.Name).
			SetSize(snap.Size).
			SetVolumeSize(snap.VolumeSize).
			SetStatus(snap.Status).
			SetCreatedAtAPI(snap.CreatedAtAPI).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("create snapshot %s: %w", snap.Name, err)
		}
	}
	return nil
}

// DeleteStaleBlockVolumes removes block volumes not collected in the latest run for the given region.
func (s *Service) DeleteStaleBlockVolumes(ctx context.Context, projectID, region string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGreenNodeVolumeBlockVolume.Query().
		Where(
			bronzegreennodevolumeblockvolume.ProjectID(projectID),
			bronzegreennodevolumeblockvolume.Region(region),
			bronzegreennodevolumeblockvolume.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale block volumes: %w", err)
	}

	for _, vol := range stale {
		if err := s.history.CloseHistory(ctx, tx, vol.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for block volume %s: %w", vol.ID, err)
		}
		if err := s.deleteBlockVolumeChildren(ctx, tx, vol.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete children for block volume %s: %w", vol.ID, err)
		}
		if err := tx.BronzeGreenNodeVolumeBlockVolume.DeleteOneID(vol.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete block volume %s: %w", vol.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
