package vpc

import (
	"context"
	"fmt"
	"time"

	entnet "danny.vn/hotpot/pkg/storage/ent/greennode/network"
	"danny.vn/hotpot/pkg/storage/ent/greennode/network/bronzehistorygreennodenetworkvpc"
)

// HistoryService handles history tracking for VPCs.
type HistoryService struct {
	entClient *entnet.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entnet.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new VPC.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entnet.Tx, data *VPCData, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeNetworkVpc.Create().
		SetResourceID(data.UUID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetCidr(data.Cidr).
		SetStatus(data.Status).
		SetRouteTableID(data.RouteTableID).
		SetRouteTableName(data.RouteTableName).
		SetDhcpOptionID(data.DhcpOptionID).
		SetDhcpOptionName(data.DhcpOptionName).
		SetDNSStatus(data.DnsStatus).
		SetDNSID(data.DnsID).
		SetZoneUUID(data.ZoneUuid).
		SetZoneName(data.ZoneName).
		SetCreatedAt(data.CreatedAt).
		SetElasticIps(data.ElasticIps).
		SetRegion(data.Region).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create vpc history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entnet.Tx, old *entnet.BronzeGreenNodeNetworkVpc, new *VPCData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeNetworkVpc.Query().
		Where(
			bronzehistorygreennodenetworkvpc.ResourceID(old.ID),
			bronzehistorygreennodenetworkvpc.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current vpc history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeNetworkVpc.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close vpc history: %w", err)
	}

	_, err = tx.BronzeHistoryGreenNodeNetworkVpc.Create().
		SetResourceID(new.UUID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetCidr(new.Cidr).
		SetStatus(new.Status).
		SetRouteTableID(new.RouteTableID).
		SetRouteTableName(new.RouteTableName).
		SetDhcpOptionID(new.DhcpOptionID).
		SetDhcpOptionName(new.DhcpOptionName).
		SetDNSStatus(new.DnsStatus).
		SetDNSID(new.DnsID).
		SetZoneUUID(new.ZoneUuid).
		SetZoneName(new.ZoneName).
		SetCreatedAt(new.CreatedAt).
		SetElasticIps(new.ElasticIps).
		SetRegion(new.Region).
		SetProjectID(new.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new vpc history: %w", err)
	}
	return nil
}

// CloseHistory closes history for a deleted VPC.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entnet.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeNetworkVpc.Query().
		Where(
			bronzehistorygreennodenetworkvpc.ResourceID(resourceID),
			bronzehistorygreennodenetworkvpc.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entnet.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current vpc history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeNetworkVpc.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close vpc history: %w", err)
	}
	return nil
}
