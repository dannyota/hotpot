package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzes1agent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzes1agentnic"
)

// Service handles SentinelOne agent ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new agent ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of agent ingestion.
type IngestResult struct {
	AgentCount     int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches all agents from SentinelOne using cursor pagination.
func (s *Service) Ingest(ctx context.Context, heartbeat func()) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	var allAgents []*AgentData
	cursor := ""

	for {
		batch, err := s.client.GetAgentsBatch(cursor)
		if err != nil {
			return nil, fmt.Errorf("get agents batch: %w", err)
		}

		for _, apiAgent := range batch.Agents {
			data, err := ConvertAgent(apiAgent, collectedAt)
			if err != nil {
				return nil, fmt.Errorf("convert agent %s: %w", apiAgent.ID, err)
			}
			allAgents = append(allAgents, data)
		}

		if heartbeat != nil {
			heartbeat()
		}

		if !batch.HasMore {
			break
		}
		cursor = batch.NextCursor
	}

	if err := s.saveAgents(ctx, allAgents); err != nil {
		return nil, fmt.Errorf("save agents: %w", err)
	}

	return &IngestResult{
		AgentCount:     len(allAgents),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveAgents(ctx context.Context, agents []*AgentData) error {
	if len(agents) == 0 {
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

	for _, data := range agents {
		existing, err := tx.BronzeS1Agent.Query().
			Where(bronzes1agent.ID(data.ResourceID)).
			WithNics().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing agent %s: %w", data.ResourceID, err)
		}

		diff := DiffAgentData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeS1Agent.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for agent %s: %w", data.ResourceID, err)
			}
			continue
		}

		// Delete old children if updating
		if existing != nil {
			if err := s.deleteAgentChildren(ctx, tx, data.ResourceID); err != nil {
				tx.Rollback()
				return fmt.Errorf("delete old children for agent %s: %w", data.ResourceID, err)
			}
		}

		var savedAgent *ent.BronzeS1Agent
		if existing == nil {
			create := tx.BronzeS1Agent.Create().
				SetID(data.ResourceID).
				SetComputerName(data.ComputerName).
				SetExternalIP(data.ExternalIP).
				SetSiteName(data.SiteName).
				SetAccountID(data.AccountID).
				SetAccountName(data.AccountName).
				SetAgentVersion(data.AgentVersion).
				SetOsType(data.OSType).
				SetOsName(data.OSName).
				SetOsRevision(data.OSRevision).
				SetOsArch(data.OSArch).
				SetIsActive(data.IsActive).
				SetIsInfected(data.IsInfected).
				SetIsDecommissioned(data.IsDecommissioned).
				SetMachineType(data.MachineType).
				SetDomain(data.Domain).
				SetUUID(data.UUID).
				SetNetworkStatus(data.NetworkStatus).
				SetActiveThreats(data.ActiveThreats).
				SetEncryptedApplications(data.EncryptedApplications).
				SetGroupName(data.GroupName).
				SetGroupID(data.GroupID).
				SetCPUCount(data.CPUCount).
				SetCoreCount(data.CoreCount).
				SetCPUID(data.CPUId).
				SetTotalMemory(data.TotalMemory).
				SetModelName(data.ModelName).
				SetSerialNumber(data.SerialNumber).
				SetStorageEncryptionStatus(data.StorageEncryptionStatus).
				SetSiteID(data.SiteID).
				SetOsUsername(data.OSUsername).
				SetGroupIP(data.GroupIP).
				SetScanStatus(data.ScanStatus).
				SetMitigationMode(data.MitigationMode).
				SetMitigationModeSuspicious(data.MitigationModeSuspicious).
				SetLastLoggedInUserName(data.LastLoggedInUserName).
				SetInstallerType(data.InstallerType).
				SetExternalID(data.ExternalID).
				SetLastIPToMgmt(data.LastIpToMgmt).
				SetIsUpToDate(data.IsUpToDate).
				SetIsPendingUninstall(data.IsPendingUninstall).
				SetIsUninstalled(data.IsUninstalled).
				SetAppsVulnerabilityStatus(data.AppsVulnerabilityStatus).
				SetConsoleMigrationStatus(data.ConsoleMigrationStatus).
				SetRangerVersion(data.RangerVersion).
				SetRangerStatus(data.RangerStatus).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.LastActiveDate != nil {
				create.SetLastActiveDate(*data.LastActiveDate)
			}
			if data.RegisteredAt != nil {
				create.SetRegisteredAt(*data.RegisteredAt)
			}
			if data.APIUpdatedAt != nil {
				create.SetAPIUpdatedAt(*data.APIUpdatedAt)
			}
			if data.OSStartTime != nil {
				create.SetOsStartTime(*data.OSStartTime)
			}
			if data.NetworkInterfacesJSON != nil {
				create.SetNetworkInterfacesJSON(data.NetworkInterfacesJSON)
			}
			if data.APICreatedAt != nil {
				create.SetAPICreatedAt(*data.APICreatedAt)
			}
			if data.ScanStartedAt != nil {
				create.SetScanStartedAt(*data.ScanStartedAt)
			}
			if data.ScanFinishedAt != nil {
				create.SetScanFinishedAt(*data.ScanFinishedAt)
			}
			if data.ActiveDirectoryJSON != nil {
				create.SetActiveDirectoryJSON(data.ActiveDirectoryJSON)
			}
			if data.LocationsJSON != nil {
				create.SetLocationsJSON(data.LocationsJSON)
			}
			if data.UserActionsNeededJSON != nil {
				create.SetUserActionsNeededJSON(data.UserActionsNeededJSON)
			}
			if data.MissingPermissionsJSON != nil {
				create.SetMissingPermissionsJSON(data.MissingPermissionsJSON)
			}

			savedAgent, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create agent %s: %w", data.ResourceID, err)
			}
		} else {
			update := tx.BronzeS1Agent.UpdateOneID(data.ResourceID).
				SetComputerName(data.ComputerName).
				SetExternalIP(data.ExternalIP).
				SetSiteName(data.SiteName).
				SetAccountID(data.AccountID).
				SetAccountName(data.AccountName).
				SetAgentVersion(data.AgentVersion).
				SetOsType(data.OSType).
				SetOsName(data.OSName).
				SetOsRevision(data.OSRevision).
				SetOsArch(data.OSArch).
				SetIsActive(data.IsActive).
				SetIsInfected(data.IsInfected).
				SetIsDecommissioned(data.IsDecommissioned).
				SetMachineType(data.MachineType).
				SetDomain(data.Domain).
				SetUUID(data.UUID).
				SetNetworkStatus(data.NetworkStatus).
				SetActiveThreats(data.ActiveThreats).
				SetEncryptedApplications(data.EncryptedApplications).
				SetGroupName(data.GroupName).
				SetGroupID(data.GroupID).
				SetCPUCount(data.CPUCount).
				SetCoreCount(data.CoreCount).
				SetCPUID(data.CPUId).
				SetTotalMemory(data.TotalMemory).
				SetModelName(data.ModelName).
				SetSerialNumber(data.SerialNumber).
				SetStorageEncryptionStatus(data.StorageEncryptionStatus).
				SetSiteID(data.SiteID).
				SetOsUsername(data.OSUsername).
				SetGroupIP(data.GroupIP).
				SetScanStatus(data.ScanStatus).
				SetMitigationMode(data.MitigationMode).
				SetMitigationModeSuspicious(data.MitigationModeSuspicious).
				SetLastLoggedInUserName(data.LastLoggedInUserName).
				SetInstallerType(data.InstallerType).
				SetExternalID(data.ExternalID).
				SetLastIPToMgmt(data.LastIpToMgmt).
				SetIsUpToDate(data.IsUpToDate).
				SetIsPendingUninstall(data.IsPendingUninstall).
				SetIsUninstalled(data.IsUninstalled).
				SetAppsVulnerabilityStatus(data.AppsVulnerabilityStatus).
				SetConsoleMigrationStatus(data.ConsoleMigrationStatus).
				SetRangerVersion(data.RangerVersion).
				SetRangerStatus(data.RangerStatus).
				SetCollectedAt(data.CollectedAt)

			if data.LastActiveDate != nil {
				update.SetLastActiveDate(*data.LastActiveDate)
			}
			if data.RegisteredAt != nil {
				update.SetRegisteredAt(*data.RegisteredAt)
			}
			if data.APIUpdatedAt != nil {
				update.SetAPIUpdatedAt(*data.APIUpdatedAt)
			}
			if data.OSStartTime != nil {
				update.SetOsStartTime(*data.OSStartTime)
			}
			if data.NetworkInterfacesJSON != nil {
				update.SetNetworkInterfacesJSON(data.NetworkInterfacesJSON)
			}
			if data.APICreatedAt != nil {
				update.SetAPICreatedAt(*data.APICreatedAt)
			}
			if data.ScanStartedAt != nil {
				update.SetScanStartedAt(*data.ScanStartedAt)
			}
			if data.ScanFinishedAt != nil {
				update.SetScanFinishedAt(*data.ScanFinishedAt)
			}
			if data.ActiveDirectoryJSON != nil {
				update.SetActiveDirectoryJSON(data.ActiveDirectoryJSON)
			}
			if data.LocationsJSON != nil {
				update.SetLocationsJSON(data.LocationsJSON)
			}
			if data.UserActionsNeededJSON != nil {
				update.SetUserActionsNeededJSON(data.UserActionsNeededJSON)
			}
			if data.MissingPermissionsJSON != nil {
				update.SetMissingPermissionsJSON(data.MissingPermissionsJSON)
			}

			savedAgent, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update agent %s: %w", data.ResourceID, err)
			}
		}

		// Create child NICs
		if err := s.createAgentChildren(ctx, tx, savedAgent, data); err != nil {
			tx.Rollback()
			return fmt.Errorf("create children for agent %s: %w", data.ResourceID, err)
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for agent %s: %w", data.ResourceID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, data, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for agent %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (s *Service) deleteAgentChildren(ctx context.Context, tx *ent.Tx, agentID string) error {
	_, err := tx.BronzeS1AgentNIC.Delete().
		Where(bronzes1agentnic.HasAgentWith(bronzes1agent.ID(agentID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete NICs: %w", err)
	}
	return nil
}

func (s *Service) createAgentChildren(ctx context.Context, tx *ent.Tx, agent *ent.BronzeS1Agent, data *AgentData) error {
	for _, nic := range data.NICs {
		nicCreate := tx.BronzeS1AgentNIC.Create().
			SetAgent(agent).
			SetInterfaceID(nic.InterfaceID).
			SetName(nic.Name).
			SetDescription(nic.Description).
			SetType(nic.Type).
			SetPhysical(nic.Physical).
			SetGatewayIP(nic.GatewayIP).
			SetGatewayMAC(nic.GatewayMac)

		if nic.InetJSON != nil {
			nicCreate.SetInetJSON(nic.InetJSON)
		}
		if nic.Inet6JSON != nil {
			nicCreate.SetInet6JSON(nic.Inet6JSON)
		}

		if _, err := nicCreate.Save(ctx); err != nil {
			return fmt.Errorf("create NIC: %w", err)
		}
	}
	return nil
}

// DeleteStale removes agents that were not collected in the latest run.
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

	stale, err := tx.BronzeS1Agent.Query().
		Where(bronzes1agent.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, agent := range stale {
		if err := s.history.CloseHistory(ctx, tx, agent.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for agent %s: %w", agent.ID, err)
		}

		if err := s.deleteAgentChildren(ctx, tx, agent.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete children for agent %s: %w", agent.ID, err)
		}

		if err := tx.BronzeS1Agent.DeleteOne(agent).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete agent %s: %w", agent.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
