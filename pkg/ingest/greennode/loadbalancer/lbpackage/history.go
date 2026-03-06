package lbpackage

import (
	"context"
	"fmt"
	"time"

	entlb "danny.vn/hotpot/pkg/storage/ent/greennode/loadbalancer"
	"danny.vn/hotpot/pkg/storage/ent/greennode/loadbalancer/bronzehistorygreennodeloadbalancerpackage"
)

// HistoryService handles history tracking for load balancer packages.
type HistoryService struct {
	entClient *entlb.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entlb.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new package.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entlb.Tx, data *PackageData, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeLoadBalancerPackage.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetType(data.Type).
		SetConnectionNumber(data.ConnectionNumber).
		SetDataTransfer(data.DataTransfer).
		SetMode(data.Mode).
		SetLbType(data.LbType).
		SetDisplayLbType(data.DisplayLbType).
		SetRegion(data.Region).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create package history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entlb.Tx, old *entlb.BronzeGreenNodeLoadBalancerPackage, new *PackageData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeLoadBalancerPackage.Query().
		Where(
			bronzehistorygreennodeloadbalancerpackage.ResourceID(old.ID),
			bronzehistorygreennodeloadbalancerpackage.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current package history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeLoadBalancerPackage.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close package history: %w", err)
	}

	_, err = tx.BronzeHistoryGreenNodeLoadBalancerPackage.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetType(new.Type).
		SetConnectionNumber(new.ConnectionNumber).
		SetDataTransfer(new.DataTransfer).
		SetMode(new.Mode).
		SetLbType(new.LbType).
		SetDisplayLbType(new.DisplayLbType).
		SetRegion(new.Region).
		SetProjectID(new.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new package history: %w", err)
	}
	return nil
}

// CloseHistory closes history for a deleted package.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entlb.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeLoadBalancerPackage.Query().
		Where(
			bronzehistorygreennodeloadbalancerpackage.ResourceID(resourceID),
			bronzehistorygreennodeloadbalancerpackage.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entlb.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current package history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeLoadBalancerPackage.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close package history: %w", err)
	}
	return nil
}
