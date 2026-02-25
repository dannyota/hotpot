package endpoint

import (
	"context"
	"fmt"
	"time"

	entnet "github.com/dannyota/hotpot/pkg/storage/ent/greennode/network"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/network/bronzegreennodenetworkendpoint"
)

// Service handles GreenNode endpoint ingestion.
type Service struct {
	client    *Client
	entClient *entnet.Client
	history   *HistoryService
}

// NewService creates a new endpoint ingestion service.
func NewService(client *Client, entClient *entnet.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of endpoint ingestion.
type IngestResult struct {
	EndpointCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches endpoints from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID, region string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	endpoints, err := s.client.ListEndpoints(ctx)
	if err != nil {
		return nil, fmt.Errorf("list endpoints: %w", err)
	}

	dataList := make([]*EndpointData, 0, len(endpoints))
	for _, e := range endpoints {
		dataList = append(dataList, ConvertEndpoint(e, projectID, region, collectedAt))
	}

	if err := s.saveEndpoints(ctx, dataList); err != nil {
		return nil, fmt.Errorf("save endpoints: %w", err)
	}

	return &IngestResult{
		EndpointCount:  len(dataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveEndpoints(ctx context.Context, endpoints []*EndpointData) error {
	if len(endpoints) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, data := range endpoints {
		existing, err := tx.BronzeGreenNodeNetworkEndpoint.Query().
			Where(bronzegreennodenetworkendpoint.ID(data.UUID)).
			First(ctx)
		if err != nil && !entnet.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing endpoint %s: %w", data.Name, err)
		}

		diff := DiffEndpointData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodeNetworkEndpoint.UpdateOneID(data.UUID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for endpoint %s: %w", data.Name, err)
			}
			continue
		}

		if existing == nil {
			_, err = tx.BronzeGreenNodeNetworkEndpoint.Create().
				SetID(data.UUID).
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
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create endpoint %s: %w", data.Name, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for endpoint %s: %w", data.Name, err)
			}
		} else {
			_, err = tx.BronzeGreenNodeNetworkEndpoint.UpdateOneID(data.UUID).
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
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update endpoint %s: %w", data.Name, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for endpoint %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// DeleteStaleEndpoints removes endpoints not collected in the latest run for the given region.
func (s *Service) DeleteStaleEndpoints(ctx context.Context, projectID, region string, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	stale, err := tx.BronzeGreenNodeNetworkEndpoint.Query().
		Where(
			bronzegreennodenetworkendpoint.ProjectID(projectID),
			bronzegreennodenetworkendpoint.Region(region),
			bronzegreennodenetworkendpoint.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale endpoints: %w", err)
	}

	for _, e := range stale {
		if err := s.history.CloseHistory(ctx, tx, e.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for endpoint %s: %w", e.ID, err)
		}
		if err := tx.BronzeGreenNodeNetworkEndpoint.DeleteOneID(e.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete endpoint %s: %w", e.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
