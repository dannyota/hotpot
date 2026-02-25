package peering

import (
	"context"
	"fmt"
	"time"

	entnet "github.com/dannyota/hotpot/pkg/storage/ent/greennode/network"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/network/bronzegreennodenetworkpeering"
)

// Service handles GreenNode peering ingestion.
type Service struct {
	client    *Client
	entClient *entnet.Client
	history   *HistoryService
}

// NewService creates a new peering ingestion service.
func NewService(client *Client, entClient *entnet.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of peering ingestion.
type IngestResult struct {
	PeeringCount   int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches peerings from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID, region string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	peerings, err := s.client.ListPeerings(ctx)
	if err != nil {
		return nil, fmt.Errorf("list peerings: %w", err)
	}

	dataList := make([]*PeeringData, 0, len(peerings))
	for _, p := range peerings {
		dataList = append(dataList, ConvertPeering(p, projectID, region, collectedAt))
	}

	if err := s.savePeerings(ctx, dataList); err != nil {
		return nil, fmt.Errorf("save peerings: %w", err)
	}

	return &IngestResult{
		PeeringCount:   len(dataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) savePeerings(ctx context.Context, peerings []*PeeringData) error {
	if len(peerings) == 0 {
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

	for _, data := range peerings {
		existing, err := tx.BronzeGreenNodeNetworkPeering.Query().
			Where(bronzegreennodenetworkpeering.ID(data.UUID)).
			First(ctx)
		if err != nil && !entnet.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing peering %s: %w", data.Name, err)
		}

		diff := DiffPeeringData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodeNetworkPeering.UpdateOneID(data.UUID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for peering %s: %w", data.Name, err)
			}
			continue
		}

		if existing == nil {
			_, err = tx.BronzeGreenNodeNetworkPeering.Create().
				SetID(data.UUID).
				SetName(data.Name).
				SetStatus(data.Status).
				SetFromVpcID(data.FromVpcID).
				SetFromCidr(data.FromCidr).
				SetEndVpcID(data.EndVpcID).
				SetEndCidr(data.EndCidr).
				SetCreatedAt(data.CreatedAt).
				SetRegion(data.Region).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create peering %s: %w", data.Name, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for peering %s: %w", data.Name, err)
			}
		} else {
			_, err = tx.BronzeGreenNodeNetworkPeering.UpdateOneID(data.UUID).
				SetName(data.Name).
				SetStatus(data.Status).
				SetFromVpcID(data.FromVpcID).
				SetFromCidr(data.FromCidr).
				SetEndVpcID(data.EndVpcID).
				SetEndCidr(data.EndCidr).
				SetCreatedAt(data.CreatedAt).
				SetRegion(data.Region).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update peering %s: %w", data.Name, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for peering %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// DeleteStalePeerings removes peerings not collected in the latest run for the given region.
func (s *Service) DeleteStalePeerings(ctx context.Context, projectID, region string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGreenNodeNetworkPeering.Query().
		Where(
			bronzegreennodenetworkpeering.ProjectID(projectID),
			bronzegreennodenetworkpeering.Region(region),
			bronzegreennodenetworkpeering.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale peerings: %w", err)
	}

	for _, p := range stale {
		if err := s.history.CloseHistory(ctx, tx, p.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for peering %s: %w", p.ID, err)
		}
		if err := tx.BronzeGreenNodeNetworkPeering.DeleteOneID(p.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete peering %s: %w", p.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
