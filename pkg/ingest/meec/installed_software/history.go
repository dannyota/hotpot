package installed_software

import (
	"context"
	"fmt"
	"time"

	entinventory "github.com/dannyota/hotpot/pkg/storage/ent/meec/inventory"
	"github.com/dannyota/hotpot/pkg/storage/ent/meec/inventory/bronzehistorymeecinventoryinstalledsoftware"
)

// HistoryService handles history tracking for installed software.
type HistoryService struct {
	entClient *entinventory.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entinventory.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) buildCreate(tx *entinventory.Tx, data *InstalledSoftwareData) *entinventory.BronzeHistoryMEECInventoryInstalledSoftwareCreate {
	return tx.BronzeHistoryMEECInventoryInstalledSoftware.Create().
		SetResourceID(data.ResourceID).
		SetComputerResourceID(data.ComputerResourceID).
		SetSoftwareID(data.SoftwareID).
		SetSoftwareName(data.SoftwareName).
		SetSoftwareVersion(data.SoftwareVersion).
		SetDisplayName(data.DisplayName).
		SetManufacturerName(data.ManufacturerName).
		SetInstalledDate(data.InstalledDate).
		SetArchitecture(data.Architecture).
		SetLocation(data.Location).
		SetSwType(data.SwType).
		SetSwCategoryName(data.SwCategoryName).
		SetDetectedTime(data.DetectedTime)
}

// CreateHistory creates a history record for a new installed software.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entinventory.Tx, data *InstalledSoftwareData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create installed software history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history for a changed installed software.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entinventory.Tx, old *entinventory.BronzeMEECInventoryInstalledSoftware, new *InstalledSoftwareData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryMEECInventoryInstalledSoftware.Query().
		Where(
			bronzehistorymeecinventoryinstalledsoftware.ResourceID(old.ID),
			bronzehistorymeecinventoryinstalledsoftware.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current installed software history: %w", err)
	}

	if err := tx.BronzeHistoryMEECInventoryInstalledSoftware.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close installed software history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new installed software history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted installed software.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entinventory.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryMEECInventoryInstalledSoftware.Query().
		Where(
			bronzehistorymeecinventoryinstalledsoftware.ResourceID(resourceID),
			bronzehistorymeecinventoryinstalledsoftware.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entinventory.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current installed software history: %w", err)
	}

	if err := tx.BronzeHistoryMEECInventoryInstalledSoftware.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close installed software history: %w", err)
	}

	return nil
}
