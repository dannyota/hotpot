package app

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorys1app"
)

// HistoryService handles history tracking for apps.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) buildCreate(tx *ent.Tx, data *AppData) *ent.BronzeHistoryS1AppCreate {
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
		SetSigned(data.Signed)

	if data.InstalledDate != nil {
		create.SetInstalledDate(*data.InstalledDate)
	}

	return create
}

// CreateHistory creates a history record for a new app.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *AppData, now time.Time) error {
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
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeS1App, new *AppData, now time.Time) error {
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
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1App.Query().
		Where(
			bronzehistorys1app.ResourceID(resourceID),
			bronzehistorys1app.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
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
