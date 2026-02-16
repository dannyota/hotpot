package droplet

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorydodroplet"
)

// HistoryService handles history tracking for Droplets.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) buildCreate(tx *ent.Tx, data *DropletData) *ent.BronzeHistoryDODropletCreate {
	return tx.BronzeHistoryDODroplet.Create().
		SetResourceID(data.ResourceID).
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
		SetSnapshotIdsJSON(data.SnapshotIdsJSON)
}

// CreateHistory creates a history record for a new Droplet.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *DropletData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create droplet history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new for a changed Droplet.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeDODroplet, new *DropletData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDODroplet.Query().
		Where(
			bronzehistorydodroplet.ResourceID(old.ID),
			bronzehistorydodroplet.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current droplet history: %w", err)
	}

	if err := tx.BronzeHistoryDODroplet.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close droplet history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new droplet history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted Droplet.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDODroplet.Query().
		Where(
			bronzehistorydodroplet.ResourceID(resourceID),
			bronzehistorydodroplet.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current droplet history: %w", err)
	}

	if err := tx.BronzeHistoryDODroplet.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close droplet history: %w", err)
	}

	return nil
}
