package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorys1agent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorys1agentnic"
)

// HistoryService handles history tracking for agents.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new agent and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *AgentData, now time.Time) error {
	agentHistCreate := h.buildAgentHistoryCreate(tx, data, data.CollectedAt, now)

	agentHist, err := agentHistCreate.Save(ctx)
	if err != nil {
		return fmt.Errorf("create agent history: %w", err)
	}

	return h.createNICsHistory(ctx, tx, agentHist.HistoryID, data.NICs, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeS1Agent, new *AgentData, diff *AgentDiff, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1Agent.Query().
		Where(
			bronzehistorys1agent.ResourceID(old.ID),
			bronzehistorys1agent.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current agent history: %w", err)
	}

	if diff.IsChanged {
		if err := tx.BronzeHistoryS1Agent.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx); err != nil {
			return fmt.Errorf("close agent history: %w", err)
		}

		agentHistCreate := h.buildAgentHistoryCreate(tx, new, old.FirstCollectedAt, now)
		agentHist, err := agentHistCreate.Save(ctx)
		if err != nil {
			return fmt.Errorf("create new agent history: %w", err)
		}

		if err := h.closeNICsHistory(ctx, tx, currentHist.HistoryID, now); err != nil {
			return fmt.Errorf("close NICs history: %w", err)
		}
		return h.createNICsHistory(ctx, tx, agentHist.HistoryID, new.NICs, now)
	}

	// Agent unchanged, check children
	if diff.NICsDiff.Changed {
		if err := h.closeNICsHistory(ctx, tx, currentHist.HistoryID, now); err != nil {
			return err
		}
		return h.createNICsHistory(ctx, tx, currentHist.HistoryID, new.NICs, now)
	}

	return nil
}

// CloseHistory closes history records for a deleted agent.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1Agent.Query().
		Where(
			bronzehistorys1agent.ResourceID(resourceID),
			bronzehistorys1agent.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current agent history: %w", err)
	}

	if err := tx.BronzeHistoryS1Agent.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close agent history: %w", err)
	}

	return h.closeNICsHistory(ctx, tx, currentHist.HistoryID, now)
}

func (h *HistoryService) buildAgentHistoryCreate(tx *ent.Tx, data *AgentData, firstCollectedAt time.Time, now time.Time) *ent.BronzeHistoryS1AgentCreate {
	create := tx.BronzeHistoryS1Agent.Create().
		SetResourceID(data.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(firstCollectedAt).
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
		SetThreatCount(data.ThreatCount).
		SetEncryptedApplications(data.EncryptedApplications).
		SetGroupName(data.GroupName).
		SetGroupID(data.GroupID).
		SetCPUCount(data.CPUCount).
		SetCoreCount(data.CoreCount).
		SetCPUID(data.CPUId).
		SetTotalMemory(data.TotalMemory).
		SetModelName(data.ModelName).
		SetSerialNumber(data.SerialNumber).
		SetStorageEncryptionStatus(data.StorageEncryptionStatus)

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

	return create
}

func (h *HistoryService) createNICsHistory(ctx context.Context, tx *ent.Tx, agentHistoryID uint, nics []NICData, now time.Time) error {
	for _, nic := range nics {
		nicCreate := tx.BronzeHistoryS1AgentNIC.Create().
			SetAgentHistoryID(agentHistoryID).
			SetValidFrom(now).
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
			return fmt.Errorf("create NIC history: %w", err)
		}
	}
	return nil
}

func (h *HistoryService) closeNICsHistory(ctx context.Context, tx *ent.Tx, agentHistoryID uint, now time.Time) error {
	_, err := tx.BronzeHistoryS1AgentNIC.Update().
		Where(
			bronzehistorys1agentnic.AgentHistoryID(agentHistoryID),
			bronzehistorys1agentnic.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close NIC history: %w", err)
	}
	return nil
}
