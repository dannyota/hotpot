package lbpackage

import (
	"context"
	"fmt"
	"time"

	entlb "danny.vn/hotpot/pkg/storage/ent/greennode/loadbalancer"
	"danny.vn/hotpot/pkg/storage/ent/greennode/loadbalancer/bronzegreennodeloadbalancerpackage"
)

// Service handles GreenNode load balancer package ingestion.
type Service struct {
	client    *Client
	entClient *entlb.Client
	history   *HistoryService
}

// NewService creates a new package ingestion service.
func NewService(client *Client, entClient *entlb.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of package ingestion.
type IngestResult struct {
	PackageCount   int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches packages from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID, region string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	pkgs, err := s.client.ListPackages(ctx)
	if err != nil {
		return nil, fmt.Errorf("list packages: %w", err)
	}

	dataList := make([]*PackageData, 0, len(pkgs))
	for _, p := range pkgs {
		dataList = append(dataList, ConvertPackage(p, projectID, region, collectedAt))
	}

	if err := s.savePackages(ctx, dataList); err != nil {
		return nil, fmt.Errorf("save packages: %w", err)
	}

	return &IngestResult{
		PackageCount:   len(dataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) savePackages(ctx context.Context, pkgs []*PackageData) error {
	if len(pkgs) == 0 {
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

	for _, data := range pkgs {
		existing, err := tx.BronzeGreenNodeLoadBalancerPackage.Query().
			Where(bronzegreennodeloadbalancerpackage.ID(data.ID)).
			First(ctx)
		if err != nil && !entlb.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing package %s: %w", data.Name, err)
		}

		diff := DiffPackageData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodeLoadBalancerPackage.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for package %s: %w", data.Name, err)
			}
			continue
		}

		if existing == nil {
			_, err = tx.BronzeGreenNodeLoadBalancerPackage.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetType(data.Type).
				SetConnectionNumber(data.ConnectionNumber).
				SetDataTransfer(data.DataTransfer).
				SetMode(data.Mode).
				SetLbType(data.LbType).
				SetDisplayLbType(data.DisplayLbType).
				SetRegion(data.Region).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create package %s: %w", data.Name, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for package %s: %w", data.Name, err)
			}
		} else {
			_, err = tx.BronzeGreenNodeLoadBalancerPackage.UpdateOneID(data.ID).
				SetName(data.Name).
				SetType(data.Type).
				SetConnectionNumber(data.ConnectionNumber).
				SetDataTransfer(data.DataTransfer).
				SetMode(data.Mode).
				SetLbType(data.LbType).
				SetDisplayLbType(data.DisplayLbType).
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update package %s: %w", data.Name, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for package %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// DeleteStalePackages removes packages not collected in the latest run.
func (s *Service) DeleteStalePackages(ctx context.Context, projectID, region string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGreenNodeLoadBalancerPackage.Query().
		Where(
			bronzegreennodeloadbalancerpackage.ProjectID(projectID),
			bronzegreennodeloadbalancerpackage.Region(region),
			bronzegreennodeloadbalancerpackage.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale packages: %w", err)
	}

	for _, p := range stale {
		if err := s.history.CloseHistory(ctx, tx, p.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for package %s: %w", p.ID, err)
		}
		if err := tx.BronzeGreenNodeLoadBalancerPackage.DeleteOneID(p.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete package %s: %w", p.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
