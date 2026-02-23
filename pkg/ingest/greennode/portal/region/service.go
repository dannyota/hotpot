package region

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegreennodeportalregion"
)

// Service handles GreenNode region ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new region ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of region ingestion.
type IngestResult struct {
	RegionCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches regions from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	regions, err := s.client.ListRegions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list regions: %w", err)
	}

	dataList := make([]*RegionData, 0, len(regions))
	for _, r := range regions {
		dataList = append(dataList, ConvertRegion(r, projectID, collectedAt))
	}

	if err := s.saveRegions(ctx, dataList); err != nil {
		return nil, fmt.Errorf("save regions: %w", err)
	}

	return &IngestResult{
		RegionCount:    len(dataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveRegions(ctx context.Context, regions []*RegionData) error {
	if len(regions) == 0 {
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

	for _, data := range regions {
		existing, err := tx.BronzeGreenNodePortalRegion.Query().
			Where(bronzegreennodeportalregion.ID(data.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing region %s: %w", data.Name, err)
		}

		diff := DiffRegionData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodePortalRegion.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for region %s: %w", data.Name, err)
			}
			continue
		}

		if existing == nil {
			_, err = tx.BronzeGreenNodePortalRegion.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create region %s: %w", data.Name, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for region %s: %w", data.Name, err)
			}
		} else {
			_, err = tx.BronzeGreenNodePortalRegion.UpdateOneID(data.ID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update region %s: %w", data.Name, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for region %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// DeleteStaleRegions removes regions not collected in the latest run.
func (s *Service) DeleteStaleRegions(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGreenNodePortalRegion.Query().
		Where(
			bronzegreennodeportalregion.ProjectID(projectID),
			bronzegreennodeportalregion.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale regions: %w", err)
	}

	for _, r := range stale {
		if err := s.history.CloseHistory(ctx, tx, r.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for region %s: %w", r.ID, err)
		}
		if err := tx.BronzeGreenNodePortalRegion.DeleteOneID(r.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete region %s: %w", r.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
