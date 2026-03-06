package computer

import (
	"context"
	"fmt"
	"time"

	entinventory "danny.vn/hotpot/pkg/storage/ent/meec/inventory"
	"danny.vn/hotpot/pkg/storage/ent/meec/inventory/bronzehistorymeecinventorycomputer"
)

// HistoryService handles history tracking for computers.
type HistoryService struct {
	entClient *entinventory.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entinventory.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new computer.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entinventory.Tx, data *ComputerData, now time.Time) error {
	create := h.buildComputerHistoryCreate(tx, data, data.CollectedAt, now)

	if _, err := create.Save(ctx); err != nil {
		return fmt.Errorf("create computer history: %w", err)
	}

	return nil
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entinventory.Tx, old *entinventory.BronzeMEECInventoryComputer, new *ComputerData, diff *ComputerDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	currentHist, err := tx.BronzeHistoryMEECInventoryComputer.Query().
		Where(
			bronzehistorymeecinventorycomputer.ResourceID(old.ID),
			bronzehistorymeecinventorycomputer.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current computer history: %w", err)
	}

	if err := tx.BronzeHistoryMEECInventoryComputer.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close computer history: %w", err)
	}

	create := h.buildComputerHistoryCreate(tx, new, old.FirstCollectedAt, now)
	if _, err := create.Save(ctx); err != nil {
		return fmt.Errorf("create new computer history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted computer.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entinventory.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryMEECInventoryComputer.Query().
		Where(
			bronzehistorymeecinventorycomputer.ResourceID(resourceID),
			bronzehistorymeecinventorycomputer.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entinventory.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current computer history: %w", err)
	}

	if err := tx.BronzeHistoryMEECInventoryComputer.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close computer history: %w", err)
	}

	return nil
}

func (h *HistoryService) buildComputerHistoryCreate(tx *entinventory.Tx, data *ComputerData, firstCollectedAt time.Time, now time.Time) *entinventory.BronzeHistoryMEECInventoryComputerCreate {
	return tx.BronzeHistoryMEECInventoryComputer.Create().
		SetResourceID(data.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(firstCollectedAt).
		SetResourceName(data.ResourceName).
		SetFqdnName(data.FQDNName).
		SetDomainNetbiosName(data.DomainNetbiosName).
		SetIPAddress(data.IPAddress).
		SetMACAddress(data.MACAddress).
		SetOsName(data.OsName).
		SetOsPlatform(data.OsPlatform).
		SetOsPlatformName(data.OsPlatformName).
		SetOsVersion(data.OsVersion).
		SetServicePack(data.ServicePack).
		SetAgentVersion(data.AgentVersion).
		SetComputerLiveStatus(data.ComputerLiveStatus).
		SetInstallationStatus(data.InstallationStatus).
		SetManagedStatus(data.ManagedStatus).
		SetBranchOfficeName(data.BranchOfficeName).
		SetOwner(data.Owner).
		SetOwnerEmailID(data.OwnerEmailID).
		SetDescription(data.Description).
		SetLocation(data.Location).
		SetLastSyncTime(data.LastSyncTime).
		SetAgentLastContactTime(data.AgentLastContactTime).
		SetAgentInstalledOn(data.AgentInstalledOn).
		SetCustomerName(data.CustomerName).
		SetCustomerID(data.CustomerID)
}
