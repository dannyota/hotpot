package vpc

import (
	"context"
	"fmt"
	"time"

	entnet "github.com/dannyota/hotpot/pkg/storage/ent/greennode/network"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/network/bronzegreennodenetworkvpc"
)

// Service handles GreenNode VPC ingestion.
type Service struct {
	client    *Client
	entClient *entnet.Client
	history   *HistoryService
}

// NewService creates a new VPC ingestion service.
func NewService(client *Client, entClient *entnet.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of VPC ingestion.
type IngestResult struct {
	VPCCount       int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches VPCs from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID, region string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	vpcs, err := s.client.ListVPCs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list vpcs: %w", err)
	}

	dataList := make([]*VPCData, 0, len(vpcs))
	for _, n := range vpcs {
		dataList = append(dataList, ConvertVPC(n, projectID, region, collectedAt))
	}

	if err := s.saveVPCs(ctx, dataList); err != nil {
		return nil, fmt.Errorf("save vpcs: %w", err)
	}

	return &IngestResult{
		VPCCount:       len(dataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveVPCs(ctx context.Context, vpcs []*VPCData) error {
	if len(vpcs) == 0 {
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

	for _, data := range vpcs {
		existing, err := tx.BronzeGreenNodeNetworkVpc.Query().
			Where(bronzegreennodenetworkvpc.ID(data.UUID)).
			First(ctx)
		if err != nil && !entnet.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing vpc %s: %w", data.Name, err)
		}

		diff := DiffVPCData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodeNetworkVpc.UpdateOneID(data.UUID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for vpc %s: %w", data.Name, err)
			}
			continue
		}

		if existing == nil {
			_, err = tx.BronzeGreenNodeNetworkVpc.Create().
				SetID(data.UUID).
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
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create vpc %s: %w", data.Name, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for vpc %s: %w", data.Name, err)
			}
		} else {
			_, err = tx.BronzeGreenNodeNetworkVpc.UpdateOneID(data.UUID).
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
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update vpc %s: %w", data.Name, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for vpc %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// DeleteStaleVPCs removes VPCs not collected in the latest run for the given region.
func (s *Service) DeleteStaleVPCs(ctx context.Context, projectID, region string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGreenNodeNetworkVpc.Query().
		Where(
			bronzegreennodenetworkvpc.ProjectID(projectID),
			bronzegreennodenetworkvpc.Region(region),
			bronzegreennodenetworkvpc.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale vpcs: %w", err)
	}

	for _, v := range stale {
		if err := s.history.CloseHistory(ctx, tx, v.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for vpc %s: %w", v.ID, err)
		}
		if err := tx.BronzeGreenNodeNetworkVpc.DeleteOneID(v.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete vpc %s: %w", v.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
