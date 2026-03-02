package computer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	entinventory "github.com/dannyota/hotpot/pkg/storage/ent/meec/inventory"
	"github.com/dannyota/hotpot/pkg/storage/ent/meec/inventory/bronzemeecinventorycomputer"
)

// Service handles MEEC computer ingestion.
type Service struct {
	client    *Client
	entClient *entinventory.Client
	history   *HistoryService
}

// NewService creates a new computer ingestion service.
func NewService(client *Client, entClient *entinventory.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of computer ingestion.
type IngestResult struct {
	ComputerCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches all computers from MEEC and saves them.
func (s *Service) Ingest(ctx context.Context, heartbeat func()) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	apiComputers, err := s.client.GetAllComputers()
	if err != nil {
		return nil, fmt.Errorf("get computers: %w", err)
	}

	if heartbeat != nil {
		heartbeat()
	}

	var allComputers []*ComputerData
	for _, apiComputer := range apiComputers {
		data := ConvertComputer(apiComputer, collectedAt)
		allComputers = append(allComputers, data)
	}

	if err := s.saveComputers(ctx, allComputers); err != nil {
		return nil, fmt.Errorf("save computers: %w", err)
	}

	return &IngestResult{
		ComputerCount:  len(allComputers),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveComputers(ctx context.Context, computers []*ComputerData) error {
	if len(computers) == 0 {
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

	for _, data := range computers {
		existing, err := tx.BronzeMEECInventoryComputer.Query().
			Where(bronzemeecinventorycomputer.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !entinventory.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing computer %s: %w", data.ResourceID, err)
		}

		diff := DiffComputerData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeMEECInventoryComputer.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for computer %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			_, err = tx.BronzeMEECInventoryComputer.Create().
				SetID(data.ResourceID).
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
				SetCustomerID(data.CustomerID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create computer %s: %w", data.ResourceID, err)
			}
		} else {
			_, err = tx.BronzeMEECInventoryComputer.UpdateOneID(data.ResourceID).
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
				SetCustomerID(data.CustomerID).
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update computer %s: %w", data.ResourceID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for computer %s: %w", data.ResourceID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, data, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for computer %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStale removes computers that were not collected in the latest run.
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

	stale, err := tx.BronzeMEECInventoryComputer.Query().
		Where(bronzemeecinventorycomputer.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, comp := range stale {
		if err := s.history.CloseHistory(ctx, tx, comp.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for computer %s: %w", comp.ID, err)
		}

		if err := tx.BronzeMEECInventoryComputer.DeleteOne(comp).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete computer %s: %w", comp.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	slog.Info("meec computers: deleted stale", "count", len(stale))

	return nil
}
