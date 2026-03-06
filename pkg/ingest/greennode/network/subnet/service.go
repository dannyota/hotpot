package subnet

import (
	"context"
	"fmt"
	"time"

	entnet "danny.vn/hotpot/pkg/storage/ent/greennode/network"
	"danny.vn/hotpot/pkg/storage/ent/greennode/network/bronzegreennodenetworksubnet"
)

// Service handles GreenNode subnet ingestion.
type Service struct {
	client    *Client
	entClient *entnet.Client
	history   *HistoryService
}

// NewService creates a new subnet ingestion service.
func NewService(client *Client, entClient *entnet.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of subnet ingestion.
type IngestResult struct {
	SubnetCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches subnets from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID, region string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	subnets, err := s.client.ListSubnets(ctx)
	if err != nil {
		return nil, fmt.Errorf("list subnets: %w", err)
	}

	dataList := make([]*SubnetData, 0, len(subnets))
	for _, sub := range subnets {
		dataList = append(dataList, ConvertSubnet(sub, projectID, region, collectedAt))
	}

	if err := s.saveSubnets(ctx, dataList); err != nil {
		return nil, fmt.Errorf("save subnets: %w", err)
	}

	return &IngestResult{
		SubnetCount:    len(dataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveSubnets(ctx context.Context, subnets []*SubnetData) error {
	if len(subnets) == 0 {
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

	for _, data := range subnets {
		existing, err := tx.BronzeGreenNodeNetworkSubnet.Query().
			Where(bronzegreennodenetworksubnet.ID(data.UUID)).
			First(ctx)
		if err != nil && !entnet.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing subnet %s: %w", data.Name, err)
		}

		diff := DiffSubnetData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodeNetworkSubnet.UpdateOneID(data.UUID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for subnet %s: %w", data.Name, err)
			}
			continue
		}

		if existing == nil {
			_, err = tx.BronzeGreenNodeNetworkSubnet.Create().
				SetID(data.UUID).
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
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create subnet %s: %w", data.Name, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for subnet %s: %w", data.Name, err)
			}
		} else {
			_, err = tx.BronzeGreenNodeNetworkSubnet.UpdateOneID(data.UUID).
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
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update subnet %s: %w", data.Name, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for subnet %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// DeleteStaleSubnets removes subnets not collected in the latest run for the given region.
func (s *Service) DeleteStaleSubnets(ctx context.Context, projectID, region string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGreenNodeNetworkSubnet.Query().
		Where(
			bronzegreennodenetworksubnet.ProjectID(projectID),
			bronzegreennodenetworksubnet.Region(region),
			bronzegreennodenetworksubnet.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale subnets: %w", err)
	}

	for _, sub := range stale {
		if err := s.history.CloseHistory(ctx, tx, sub.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for subnet %s: %w", sub.ID, err)
		}
		if err := tx.BronzeGreenNodeNetworkSubnet.DeleteOneID(sub.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete subnet %s: %w", sub.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
