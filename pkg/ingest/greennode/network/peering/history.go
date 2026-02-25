package peering

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygreennodenetworkpeering"
)

// HistoryService handles history tracking for peerings.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new peering.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *PeeringData, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeNetworkPeering.Create().
		SetResourceID(data.UUID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetStatus(data.Status).
		SetFromVpcID(data.FromVpcID).
		SetFromCidr(data.FromCidr).
		SetEndVpcID(data.EndVpcID).
		SetEndCidr(data.EndCidr).
		SetCreatedAt(data.CreatedAt).
		SetRegion(data.Region).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create peering history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGreenNodeNetworkPeering, new *PeeringData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeNetworkPeering.Query().
		Where(
			bronzehistorygreennodenetworkpeering.ResourceID(old.ID),
			bronzehistorygreennodenetworkpeering.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current peering history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeNetworkPeering.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close peering history: %w", err)
	}

	_, err = tx.BronzeHistoryGreenNodeNetworkPeering.Create().
		SetResourceID(new.UUID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetStatus(new.Status).
		SetFromVpcID(new.FromVpcID).
		SetFromCidr(new.FromCidr).
		SetEndVpcID(new.EndVpcID).
		SetEndCidr(new.EndCidr).
		SetCreatedAt(new.CreatedAt).
		SetRegion(new.Region).
		SetProjectID(new.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new peering history: %w", err)
	}
	return nil
}

// CloseHistory closes history for a deleted peering.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeNetworkPeering.Query().
		Where(
			bronzehistorygreennodenetworkpeering.ResourceID(resourceID),
			bronzehistorygreennodenetworkpeering.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current peering history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeNetworkPeering.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close peering history: %w", err)
	}
	return nil
}
