package vpc

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorydovpc"
)

// HistoryService handles history tracking for VPCs.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) buildCreate(tx *ent.Tx, data *VpcData) *ent.BronzeHistoryDOVpcCreate {
	create := tx.BronzeHistoryDOVpc.Create().
		SetResourceID(data.ResourceID).
		SetName(data.Name).
		SetDescription(data.Description).
		SetRegion(data.Region).
		SetIPRange(data.IPRange).
		SetUrn(data.URN).
		SetIsDefault(data.IsDefault)

	if data.APICreatedAt != nil {
		create.SetAPICreatedAt(*data.APICreatedAt)
	}

	return create
}

// CreateHistory creates a history record for a new VPC.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *VpcData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create VPC history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new for a changed VPC.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeDOVpc, new *VpcData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDOVpc.Query().
		Where(
			bronzehistorydovpc.ResourceID(old.ID),
			bronzehistorydovpc.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current VPC history: %w", err)
	}

	if err := tx.BronzeHistoryDOVpc.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close VPC history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new VPC history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted VPC.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDOVpc.Query().
		Where(
			bronzehistorydovpc.ResourceID(resourceID),
			bronzehistorydovpc.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current VPC history: %w", err)
	}

	if err := tx.BronzeHistoryDOVpc.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close VPC history: %w", err)
	}

	return nil
}
