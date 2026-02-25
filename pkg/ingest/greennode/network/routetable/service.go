package routetable

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegreennodenetworkroutetable"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegreennodenetworkroutetableroute"
)

// Service handles GreenNode route table ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new route table ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of route table ingestion.
type IngestResult struct {
	RouteTableCount int
	CollectedAt     time.Time
	DurationMillis  int64
}

// Ingest fetches route tables from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID, region string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	routeTables, err := s.client.ListRouteTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("list route tables: %w", err)
	}

	dataList := make([]*RouteTableData, 0, len(routeTables))
	for _, rt := range routeTables {
		dataList = append(dataList, ConvertRouteTable(rt, projectID, region, collectedAt))
	}

	if err := s.saveRouteTables(ctx, dataList); err != nil {
		return nil, fmt.Errorf("save route tables: %w", err)
	}

	return &IngestResult{
		RouteTableCount: len(dataList),
		CollectedAt:     collectedAt,
		DurationMillis:  time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveRouteTables(ctx context.Context, routeTables []*RouteTableData) error {
	if len(routeTables) == 0 {
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

	for _, data := range routeTables {
		existing, err := tx.BronzeGreenNodeNetworkRouteTable.Query().
			Where(bronzegreennodenetworkroutetable.ID(data.UUID)).
			WithRoutes().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing route table %s: %w", data.Name, err)
		}

		diff := DiffRouteTableData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodeNetworkRouteTable.UpdateOneID(data.UUID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for route table %s: %w", data.Name, err)
			}
			continue
		}

		if existing != nil {
			if err := s.deleteRouteTableChildren(ctx, tx, data.UUID); err != nil {
				tx.Rollback()
				return fmt.Errorf("delete children for route table %s: %w", data.Name, err)
			}
		}

		var savedRouteTable *ent.BronzeGreenNodeNetworkRouteTable
		if existing == nil {
			savedRouteTable, err = tx.BronzeGreenNodeNetworkRouteTable.Create().
				SetID(data.UUID).
				SetName(data.Name).
				SetStatus(data.Status).
				SetNetworkID(data.NetworkID).
				SetCreatedAt(data.CreatedAt).
				SetRegion(data.Region).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create route table %s: %w", data.Name, err)
			}
		} else {
			savedRouteTable, err = tx.BronzeGreenNodeNetworkRouteTable.UpdateOneID(data.UUID).
				SetName(data.Name).
				SetStatus(data.Status).
				SetNetworkID(data.NetworkID).
				SetCreatedAt(data.CreatedAt).
				SetRegion(data.Region).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update route table %s: %w", data.Name, err)
			}
		}

		if err := s.createRouteTableChildren(ctx, tx, savedRouteTable, data); err != nil {
			tx.Rollback()
			return fmt.Errorf("create children for route table %s: %w", data.Name, err)
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for route table %s: %w", data.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, data, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for route table %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (s *Service) deleteRouteTableChildren(ctx context.Context, tx *ent.Tx, routeTableID string) error {
	_, err := tx.BronzeGreenNodeNetworkRouteTableRoute.Delete().
		Where(bronzegreennodenetworkroutetableroute.HasRouteTableWith(bronzegreennodenetworkroutetable.ID(routeTableID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete routes: %w", err)
	}
	return nil
}

func (s *Service) createRouteTableChildren(ctx context.Context, tx *ent.Tx, rt *ent.BronzeGreenNodeNetworkRouteTable, data *RouteTableData) error {
	for _, r := range data.Routes {
		_, err := tx.BronzeGreenNodeNetworkRouteTableRoute.Create().
			SetRouteTable(rt).
			SetRouteID(r.RouteID).
			SetRoutingType(r.RoutingType).
			SetDestinationCidrBlock(r.DestinationCidrBlock).
			SetTarget(r.Target).
			SetStatus(r.Status).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("create route %s: %w", r.RouteID, err)
		}
	}
	return nil
}

// DeleteStaleRouteTables removes route tables not collected in the latest run for the given region.
func (s *Service) DeleteStaleRouteTables(ctx context.Context, projectID, region string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGreenNodeNetworkRouteTable.Query().
		Where(
			bronzegreennodenetworkroutetable.ProjectID(projectID),
			bronzegreennodenetworkroutetable.Region(region),
			bronzegreennodenetworkroutetable.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale route tables: %w", err)
	}

	for _, rt := range stale {
		if err := s.history.CloseHistory(ctx, tx, rt.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for route table %s: %w", rt.ID, err)
		}
		if err := s.deleteRouteTableChildren(ctx, tx, rt.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete children for route table %s: %w", rt.ID, err)
		}
		if err := tx.BronzeGreenNodeNetworkRouteTable.DeleteOneID(rt.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete route table %s: %w", rt.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
