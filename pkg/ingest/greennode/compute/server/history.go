package server

import (
	"context"
	"fmt"
	"time"

	entcompute "github.com/dannyota/hotpot/pkg/storage/ent/greennode/compute"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/compute/bronzehistorygreennodecomputeserver"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/compute/bronzehistorygreennodecomputeserversecgroup"
)

// HistoryService handles history tracking for servers.
type HistoryService struct {
	entClient *entcompute.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entcompute.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new server and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entcompute.Tx, data *ServerData, now time.Time) error {
	serverHist, err := h.createServerHistory(ctx, tx, data, now, data.CollectedAt)
	if err != nil {
		return err
	}
	return h.createSecGroupsHistory(ctx, tx, serverHist.ID, data.SecGroups, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entcompute.Tx, old *entcompute.BronzeGreenNodeComputeServer, new *ServerData, diff *ServerDiff, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeComputeServer.Query().
		Where(
			bronzehistorygreennodecomputeserver.ResourceID(old.ID),
			bronzehistorygreennodecomputeserver.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current server history: %w", err)
	}

	if diff.IsChanged {
		// Close old history
		if err := tx.BronzeHistoryGreenNodeComputeServer.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx); err != nil {
			return fmt.Errorf("close server history: %w", err)
		}

		// Create new history
		serverHist, err := h.createServerHistory(ctx, tx, new, now, old.FirstCollectedAt)
		if err != nil {
			return err
		}

		// Close and recreate all children
		if err := h.closeSecGroupsHistory(ctx, tx, currentHist.ID, now); err != nil {
			return err
		}
		return h.createSecGroupsHistory(ctx, tx, serverHist.ID, new.SecGroups, now)
	}

	// Server unchanged, check children
	if diff.SecGroupsDiff.Changed {
		if err := h.closeSecGroupsHistory(ctx, tx, currentHist.ID, now); err != nil {
			return err
		}
		return h.createSecGroupsHistory(ctx, tx, currentHist.ID, new.SecGroups, now)
	}

	return nil
}

// CloseHistory closes history records for a deleted server.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entcompute.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeComputeServer.Query().
		Where(
			bronzehistorygreennodecomputeserver.ResourceID(resourceID),
			bronzehistorygreennodecomputeserver.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entcompute.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current server history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeComputeServer.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close server history: %w", err)
	}

	return h.closeSecGroupsHistory(ctx, tx, currentHist.ID, now)
}

func (h *HistoryService) createServerHistory(ctx context.Context, tx *entcompute.Tx, data *ServerData, now time.Time, firstCollectedAt time.Time) (*entcompute.BronzeHistoryGreenNodeComputeServer, error) {
	create := tx.BronzeHistoryGreenNodeComputeServer.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(firstCollectedAt).
		SetName(data.Name).
		SetStatus(data.Status).
		SetLocation(data.Location).
		SetZoneID(data.ZoneID).
		SetCreatedAtAPI(data.CreatedAtAPI).
		SetBootVolumeID(data.BootVolumeID).
		SetEncryptionVolume(data.EncryptionVolume).
		SetLicence(data.Licence).
		SetMetadata(data.Metadata).
		SetMigrateState(data.MigrateState).
		SetProduct(data.Product).
		SetServerGroupID(data.ServerGroupID).
		SetServerGroupName(data.ServerGroupName).
		SetSSHKeyName(data.SSHKeyName).
		SetStopBeforeMigrate(data.StopBeforeMigrate).
		SetUser(data.User).
		SetImageID(data.ImageID).
		SetImageType(data.ImageType).
		SetImageVersion(data.ImageVersion).
		SetFlavorID(data.FlavorID).
		SetFlavorName(data.FlavorName).
		SetFlavorCPU(data.FlavorCPU).
		SetFlavorMemory(data.FlavorMemory).
		SetFlavorGpu(data.FlavorGPU).
		SetFlavorBandwidth(data.FlavorBandwidth).
		SetRegion(data.Region).
		SetProjectID(data.ProjectID)

	if data.InterfacesJSON != nil {
		create.SetInterfacesJSON(data.InterfacesJSON)
	}

	hist, err := create.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create server history: %w", err)
	}
	return hist, nil
}

func (h *HistoryService) createSecGroupsHistory(ctx context.Context, tx *entcompute.Tx, serverHistoryID uint, secGroups []SecGroupData, now time.Time) error {
	for _, sg := range secGroups {
		_, err := tx.BronzeHistoryGreenNodeComputeServerSecGroup.Create().
			SetServerHistoryID(serverHistoryID).
			SetValidFrom(now).
			SetUUID(sg.UUID).
			SetName(sg.Name).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("create sec group history: %w", err)
		}
	}
	return nil
}

func (h *HistoryService) closeSecGroupsHistory(ctx context.Context, tx *entcompute.Tx, serverHistoryID uint, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeComputeServerSecGroup.Update().
		Where(
			bronzehistorygreennodecomputeserversecgroup.ServerHistoryID(serverHistoryID),
			bronzehistorygreennodecomputeserversecgroup.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close sec groups history: %w", err)
	}
	return nil
}
