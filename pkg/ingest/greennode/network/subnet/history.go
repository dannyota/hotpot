package subnet

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygreennodenetworksubnet"
)

// HistoryService handles history tracking for subnets.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new subnet.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *SubnetData, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeNetworkSubnet.Create().
		SetResourceID(data.UUID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetNetworkID(data.NetworkID).
		SetCidr(data.Cidr).
		SetStatus(data.Status).
		SetRouteTableID(data.RouteTableID).
		SetInterfaceACLPolicyID(data.InterfaceAclPolicyID).
		SetInterfaceACLPolicyName(data.InterfaceAclPolicyName).
		SetZoneID(data.ZoneID).
		SetSecondarySubnets(data.SecondarySubnets).
		SetRegion(data.Region).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create subnet history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGreenNodeNetworkSubnet, new *SubnetData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeNetworkSubnet.Query().
		Where(
			bronzehistorygreennodenetworksubnet.ResourceID(old.ID),
			bronzehistorygreennodenetworksubnet.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current subnet history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeNetworkSubnet.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close subnet history: %w", err)
	}

	_, err = tx.BronzeHistoryGreenNodeNetworkSubnet.Create().
		SetResourceID(new.UUID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetNetworkID(new.NetworkID).
		SetCidr(new.Cidr).
		SetStatus(new.Status).
		SetRouteTableID(new.RouteTableID).
		SetInterfaceACLPolicyID(new.InterfaceAclPolicyID).
		SetInterfaceACLPolicyName(new.InterfaceAclPolicyName).
		SetZoneID(new.ZoneID).
		SetSecondarySubnets(new.SecondarySubnets).
		SetRegion(new.Region).
		SetProjectID(new.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new subnet history: %w", err)
	}
	return nil
}

// CloseHistory closes history for a deleted subnet.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeNetworkSubnet.Query().
		Where(
			bronzehistorygreennodenetworksubnet.ResourceID(resourceID),
			bronzehistorygreennodenetworksubnet.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current subnet history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeNetworkSubnet.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close subnet history: %w", err)
	}
	return nil
}
