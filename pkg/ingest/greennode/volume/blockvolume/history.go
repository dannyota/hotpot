package blockvolume

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygreennodevolumeblockvolume"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygreennodevolumesnapshot"
)

// HistoryService handles history tracking for block volumes.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new block volume and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *BlockVolumeData, now time.Time) error {
	volHist, err := h.createBlockVolumeHistory(ctx, tx, data, now, data.CollectedAt)
	if err != nil {
		return err
	}
	return h.createSnapshotsHistory(ctx, tx, volHist.ID, data.Snapshots, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGreenNodeVolumeBlockVolume, new *BlockVolumeData, diff *BlockVolumeDiff, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeVolumeBlockVolume.Query().
		Where(
			bronzehistorygreennodevolumeblockvolume.ResourceID(old.ID),
			bronzehistorygreennodevolumeblockvolume.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current block volume history: %w", err)
	}

	if diff.IsChanged {
		// Close old history
		if err := tx.BronzeHistoryGreenNodeVolumeBlockVolume.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx); err != nil {
			return fmt.Errorf("close block volume history: %w", err)
		}

		// Create new history
		volHist, err := h.createBlockVolumeHistory(ctx, tx, new, now, old.FirstCollectedAt)
		if err != nil {
			return err
		}

		// Close and recreate all children
		if err := h.closeSnapshotsHistory(ctx, tx, currentHist.ID, now); err != nil {
			return err
		}
		return h.createSnapshotsHistory(ctx, tx, volHist.ID, new.Snapshots, now)
	}

	// Block volume unchanged, check children
	if diff.SnapshotsDiff.Changed {
		if err := h.closeSnapshotsHistory(ctx, tx, currentHist.ID, now); err != nil {
			return err
		}
		return h.createSnapshotsHistory(ctx, tx, currentHist.ID, new.Snapshots, now)
	}

	return nil
}

// CloseHistory closes history records for a deleted block volume.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeVolumeBlockVolume.Query().
		Where(
			bronzehistorygreennodevolumeblockvolume.ResourceID(resourceID),
			bronzehistorygreennodevolumeblockvolume.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current block volume history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeVolumeBlockVolume.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close block volume history: %w", err)
	}

	return h.closeSnapshotsHistory(ctx, tx, currentHist.ID, now)
}

func (h *HistoryService) createBlockVolumeHistory(ctx context.Context, tx *ent.Tx, data *BlockVolumeData, now time.Time, firstCollectedAt time.Time) (*ent.BronzeHistoryGreenNodeVolumeBlockVolume, error) {
	create := tx.BronzeHistoryGreenNodeVolumeBlockVolume.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(firstCollectedAt).
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
		SetProjectID(data.ProjectID)

	if data.AttachedMachineJSON != nil {
		create.SetAttachedMachineJSON(data.AttachedMachineJSON)
	}

	hist, err := create.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create block volume history: %w", err)
	}
	return hist, nil
}

func (h *HistoryService) createSnapshotsHistory(ctx context.Context, tx *ent.Tx, blockVolumeHistoryID uint, snapshots []SnapshotData, now time.Time) error {
	for _, s := range snapshots {
		_, err := tx.BronzeHistoryGreenNodeVolumeSnapshot.Create().
			SetBlockVolumeHistoryID(blockVolumeHistoryID).
			SetValidFrom(now).
			SetSnapshotID(s.SnapshotID).
			SetName(s.Name).
			SetSize(s.Size).
			SetVolumeSize(s.VolumeSize).
			SetStatus(s.Status).
			SetCreatedAtAPI(s.CreatedAtAPI).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("create snapshot history: %w", err)
		}
	}
	return nil
}

func (h *HistoryService) closeSnapshotsHistory(ctx context.Context, tx *ent.Tx, blockVolumeHistoryID uint, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeVolumeSnapshot.Update().
		Where(
			bronzehistorygreennodevolumesnapshot.BlockVolumeHistoryID(blockVolumeHistoryID),
			bronzehistorygreennodevolumesnapshot.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close snapshots history: %w", err)
	}
	return nil
}
