package group

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorys1group"
)

// HistoryService handles history tracking for groups.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) buildCreate(tx *ent.Tx, data *GroupData) *ent.BronzeHistoryS1GroupCreate {
	create := tx.BronzeHistoryS1Group.Create().
		SetResourceID(data.ResourceID).
		SetName(data.Name).
		SetSiteID(data.SiteID).
		SetType(data.Type).
		SetIsDefault(data.IsDefault).
		SetInherits(data.Inherits).
		SetTotalAgents(data.TotalAgents).
		SetCreator(data.Creator).
		SetCreatorID(data.CreatorID).
		SetFilterName(data.FilterName).
		SetFilterID(data.FilterID)

	if data.Rank != nil {
		create.SetRank(*data.Rank)
	}
	if data.APICreatedAt != nil {
		create.SetAPICreatedAt(*data.APICreatedAt)
	}
	if data.APIUpdatedAt != nil {
		create.SetAPIUpdatedAt(*data.APIUpdatedAt)
	}

	return create
}

// CreateHistory creates a history record for a new group.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *GroupData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create group history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new for a changed group.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeS1Group, new *GroupData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1Group.Query().
		Where(
			bronzehistorys1group.ResourceID(old.ID),
			bronzehistorys1group.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current group history: %w", err)
	}

	if err := tx.BronzeHistoryS1Group.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close group history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new group history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted group.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1Group.Query().
		Where(
			bronzehistorys1group.ResourceID(resourceID),
			bronzehistorys1group.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current group history: %w", err)
	}

	if err := tx.BronzeHistoryS1Group.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close group history: %w", err)
	}

	return nil
}
