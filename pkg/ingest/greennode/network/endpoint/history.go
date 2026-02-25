package endpoint

import (
	"context"
	"fmt"
	"time"

	entnet "github.com/dannyota/hotpot/pkg/storage/ent/greennode/network"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/network/bronzehistorygreennodenetworkendpoint"
)

// HistoryService handles history tracking for endpoints.
type HistoryService struct {
	entClient *entnet.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entnet.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new endpoint.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entnet.Tx, data *EndpointData, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeNetworkEndpoint.Create().
		SetResourceID(data.UUID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetIpv4Address(data.Ipv4Address).
		SetEndpointURL(data.EndpointURL).
		SetEndpointAuthURL(data.EndpointAuthURL).
		SetEndpointServiceID(data.EndpointServiceID).
		SetStatus(data.Status).
		SetBillingStatus(data.BillingStatus).
		SetEndpointType(data.EndpointType).
		SetVersion(data.Version).
		SetDescription(data.Description).
		SetCreatedAt(data.CreatedAt).
		SetUpdatedAt(data.UpdatedAt).
		SetVpcID(data.VpcID).
		SetVpcName(data.VpcName).
		SetZoneUUID(data.ZoneUuid).
		SetEnableDNSName(data.EnableDnsName).
		SetEndpointDomains(data.EndpointDomains).
		SetSubnetID(data.SubnetID).
		SetCategoryName(data.CategoryName).
		SetServiceName(data.ServiceName).
		SetServiceEndpointType(data.ServiceEndpointType).
		SetPackageName(data.PackageName).
		SetRegion(data.Region).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create endpoint history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entnet.Tx, old *entnet.BronzeGreenNodeNetworkEndpoint, new *EndpointData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeNetworkEndpoint.Query().
		Where(
			bronzehistorygreennodenetworkendpoint.ResourceID(old.ID),
			bronzehistorygreennodenetworkendpoint.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current endpoint history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeNetworkEndpoint.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close endpoint history: %w", err)
	}

	_, err = tx.BronzeHistoryGreenNodeNetworkEndpoint.Create().
		SetResourceID(new.UUID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetIpv4Address(new.Ipv4Address).
		SetEndpointURL(new.EndpointURL).
		SetEndpointAuthURL(new.EndpointAuthURL).
		SetEndpointServiceID(new.EndpointServiceID).
		SetStatus(new.Status).
		SetBillingStatus(new.BillingStatus).
		SetEndpointType(new.EndpointType).
		SetVersion(new.Version).
		SetDescription(new.Description).
		SetCreatedAt(new.CreatedAt).
		SetUpdatedAt(new.UpdatedAt).
		SetVpcID(new.VpcID).
		SetVpcName(new.VpcName).
		SetZoneUUID(new.ZoneUuid).
		SetEnableDNSName(new.EnableDnsName).
		SetEndpointDomains(new.EndpointDomains).
		SetSubnetID(new.SubnetID).
		SetCategoryName(new.CategoryName).
		SetServiceName(new.ServiceName).
		SetServiceEndpointType(new.ServiceEndpointType).
		SetPackageName(new.PackageName).
		SetRegion(new.Region).
		SetProjectID(new.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new endpoint history: %w", err)
	}
	return nil
}

// CloseHistory closes history for a deleted endpoint.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entnet.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeNetworkEndpoint.Query().
		Where(
			bronzehistorygreennodenetworkendpoint.ResourceID(resourceID),
			bronzehistorygreennodenetworkendpoint.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entnet.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current endpoint history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeNetworkEndpoint.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close endpoint history: %w", err)
	}
	return nil
}
