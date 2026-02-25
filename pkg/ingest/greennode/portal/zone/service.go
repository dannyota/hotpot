package zone

import (
	"context"
	"fmt"
	"time"

	entportal "github.com/dannyota/hotpot/pkg/storage/ent/greennode/portal"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/portal/bronzegreennodeportalzone"
)

// Service handles GreenNode zone ingestion.
type Service struct {
	client    *Client
	entClient *entportal.Client
	history   *HistoryService
}

// NewService creates a new zone ingestion service.
func NewService(client *Client, entClient *entportal.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of zone ingestion.
type IngestResult struct {
	ZoneCount      int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches zones from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	zones, err := s.client.ListZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("list zones: %w", err)
	}

	dataList := make([]*ZoneData, 0, len(zones))
	for _, z := range zones {
		dataList = append(dataList, ConvertZone(z, projectID, collectedAt))
	}

	if err := s.saveZones(ctx, dataList); err != nil {
		return nil, fmt.Errorf("save zones: %w", err)
	}

	return &IngestResult{
		ZoneCount:      len(dataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveZones(ctx context.Context, zones []*ZoneData) error {
	if len(zones) == 0 {
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

	for _, data := range zones {
		existing, err := tx.BronzeGreenNodePortalZone.Query().
			Where(bronzegreennodeportalzone.ID(data.ID)).
			First(ctx)
		if err != nil && !entportal.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing zone %s: %w", data.Name, err)
		}

		diff := DiffZoneData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodePortalZone.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for zone %s: %w", data.Name, err)
			}
			continue
		}

		if existing == nil {
			_, err = tx.BronzeGreenNodePortalZone.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetOpenstackZone(data.OpenstackZone).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create zone %s: %w", data.Name, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for zone %s: %w", data.Name, err)
			}
		} else {
			_, err = tx.BronzeGreenNodePortalZone.UpdateOneID(data.ID).
				SetName(data.Name).
				SetOpenstackZone(data.OpenstackZone).
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update zone %s: %w", data.Name, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for zone %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// DeleteStaleZones removes zones not collected in the latest run.
func (s *Service) DeleteStaleZones(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGreenNodePortalZone.Query().
		Where(
			bronzegreennodeportalzone.ProjectID(projectID),
			bronzegreennodeportalzone.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale zones: %w", err)
	}

	for _, z := range stale {
		if err := s.history.CloseHistory(ctx, tx, z.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for zone %s: %w", z.ID, err)
		}
		if err := tx.BronzeGreenNodePortalZone.DeleteOneID(z.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete zone %s: %w", z.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
