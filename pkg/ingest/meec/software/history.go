package software

import (
	"context"
	"fmt"
	"time"

	entinventory "danny.vn/hotpot/pkg/storage/ent/meec/inventory"
	"danny.vn/hotpot/pkg/storage/ent/meec/inventory/bronzehistorymeecinventorysoftware"
)

// HistoryService handles history tracking for software.
type HistoryService struct {
	entClient *entinventory.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entinventory.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) buildCreate(tx *entinventory.Tx, data *SoftwareData) *entinventory.BronzeHistoryMEECInventorySoftwareCreate {
	return tx.BronzeHistoryMEECInventorySoftware.Create().
		SetResourceID(data.ResourceID).
		SetSoftwareName(data.SoftwareName).
		SetSoftwareVersion(data.SoftwareVersion).
		SetDisplayName(data.DisplayName).
		SetManufacturerID(data.ManufacturerID).
		SetManufacturerName(data.ManufacturerName).
		SetSwCategoryName(data.SwCategoryName).
		SetSwType(data.SwType).
		SetSwFamily(data.SwFamily).
		SetInstalledFormat(data.InstalledFormat).
		SetIsUsageProhibited(data.IsUsageProhibited).
		SetManagedInstallations(data.ManagedInstallations).
		SetNetworkInstallations(data.NetworkInstallations).
		SetManagedSwID(data.ManagedSwID).
		SetDetectedTime(data.DetectedTime).
		SetCompliantStatus(data.CompliantStatus).
		SetTotalCopies(data.TotalCopies).
		SetRemainingCopies(data.RemainingCopies)
}

// CreateHistory creates a history record for a new software entry.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entinventory.Tx, data *SoftwareData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create software history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history for a changed software entry.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entinventory.Tx, old *entinventory.BronzeMEECInventorySoftware, new *SoftwareData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryMEECInventorySoftware.Query().
		Where(
			bronzehistorymeecinventorysoftware.ResourceID(old.ID),
			bronzehistorymeecinventorysoftware.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current software history: %w", err)
	}

	if err := tx.BronzeHistoryMEECInventorySoftware.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close software history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new software history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted software entry.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entinventory.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryMEECInventorySoftware.Query().
		Where(
			bronzehistorymeecinventorysoftware.ResourceID(resourceID),
			bronzehistorymeecinventorysoftware.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entinventory.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current software history: %w", err)
	}

	if err := tx.BronzeHistoryMEECInventorySoftware.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close software history: %w", err)
	}

	return nil
}
