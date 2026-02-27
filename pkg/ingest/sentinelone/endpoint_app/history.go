package endpoint_app

import (
	"context"
	"fmt"
	"time"

	ents1 "github.com/dannyota/hotpot/pkg/storage/ent/s1"
	"github.com/dannyota/hotpot/pkg/storage/ent/s1/bronzehistorys1endpointapp"
)

// HistoryService handles history tracking for endpoint apps.
type HistoryService struct {
	entClient *ents1.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ents1.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) buildCreate(tx *ents1.Tx, data *EndpointAppData) *ents1.BronzeHistoryS1EndpointAppCreate {
	create := tx.BronzeHistoryS1EndpointApp.Create().
		SetResourceID(data.ResourceID).
		SetAgentID(data.AgentID).
		SetName(data.Name).
		SetVersion(data.Version).
		SetPublisher(data.Publisher).
		SetSize(data.Size)

	if data.InstalledDate != nil {
		create.SetInstalledDate(*data.InstalledDate)
	}

	return create
}

// CreateHistory creates a history record for a new endpoint app.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ents1.Tx, data *EndpointAppData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create endpoint app history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history for a changed endpoint app.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ents1.Tx, old *ents1.BronzeS1EndpointApp, new *EndpointAppData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1EndpointApp.Query().
		Where(
			bronzehistorys1endpointapp.ResourceID(old.ID),
			bronzehistorys1endpointapp.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current endpoint app history: %w", err)
	}

	if err := tx.BronzeHistoryS1EndpointApp.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close endpoint app history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new endpoint app history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted endpoint app.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ents1.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1EndpointApp.Query().
		Where(
			bronzehistorys1endpointapp.ResourceID(resourceID),
			bronzehistorys1endpointapp.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ents1.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current endpoint app history: %w", err)
	}

	if err := tx.BronzeHistoryS1EndpointApp.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close endpoint app history: %w", err)
	}

	return nil
}
