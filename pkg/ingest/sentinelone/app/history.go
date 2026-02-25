package app

import (
	"context"
	"fmt"
	"time"

	ents1 "github.com/dannyota/hotpot/pkg/storage/ent/s1"
	"github.com/dannyota/hotpot/pkg/storage/ent/s1/bronzehistorys1app"
)

// HistoryService handles history tracking for apps.
type HistoryService struct {
	entClient *ents1.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ents1.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) buildCreate(tx *ents1.Tx, data *AppData) *ents1.BronzeHistoryS1AppCreate {
	create := tx.BronzeHistoryS1App.Create().
		SetResourceID(data.ResourceID).
		SetName(data.Name).
		SetPublisher(data.Publisher).
		SetVersion(data.Version).
		SetSize(data.Size).
		SetAppType(data.AppType).
		SetOsType(data.OsType).
		SetAgentID(data.AgentID).
		SetAgentComputerName(data.AgentComputerName).
		SetAgentMachineType(data.AgentMachineType).
		SetAgentIsActive(data.AgentIsActive).
		SetAgentIsDecommissioned(data.AgentIsDecommissioned).
		SetRiskLevel(data.RiskLevel).
		SetSigned(data.Signed).
		SetAgentUUID(data.AgentUUID).
		SetAgentDomain(data.AgentDomain).
		SetAgentVersion(data.AgentVersion).
		SetAgentOsType(data.AgentOsType).
		SetAgentNetworkStatus(data.AgentNetworkStatus).
		SetAgentInfected(data.AgentInfected).
		SetAgentOperationalState(data.AgentOperationalState)

	if data.InstalledDate != nil {
		create.SetInstalledDate(*data.InstalledDate)
	}
	if data.APICreatedAt != nil {
		create.SetAPICreatedAt(*data.APICreatedAt)
	}
	if data.APIUpdatedAt != nil {
		create.SetAPIUpdatedAt(*data.APIUpdatedAt)
	}

	return create
}

// CreateHistory creates a history record for a new app.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ents1.Tx, data *AppData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create app history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new for a changed app.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ents1.Tx, old *ents1.BronzeS1App, new *AppData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1App.Query().
		Where(
			bronzehistorys1app.ResourceID(old.ID),
			bronzehistorys1app.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current app history: %w", err)
	}

	if err := tx.BronzeHistoryS1App.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close app history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new app history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted app.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ents1.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1App.Query().
		Where(
			bronzehistorys1app.ResourceID(resourceID),
			bronzehistorys1app.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ents1.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current app history: %w", err)
	}

	if err := tx.BronzeHistoryS1App.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close app history: %w", err)
	}

	return nil
}
