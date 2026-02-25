package volumetypezone

import (
	"context"
	"fmt"
	"time"

	entvol "github.com/dannyota/hotpot/pkg/storage/ent/greennode/volume"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/volume/bronzegreennodevolumevolumetypezone"
)

// Service handles GreenNode volume type zone ingestion.
type Service struct {
	client    *Client
	entClient *entvol.Client
	history   *HistoryService
}

// NewService creates a new volume type zone ingestion service.
func NewService(client *Client, entClient *entvol.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of volume type zone ingestion.
type IngestResult struct {
	VolumeTypeZoneCount int
	CollectedAt         time.Time
	DurationMillis      int64
}

// Ingest fetches volume type zones from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID, region string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	zones, err := s.client.ListVolumeTypeZones(ctx)
	if err != nil {
		return nil, fmt.Errorf("list volume type zones: %w", err)
	}

	dataList := make([]*VolumeTypeZoneData, 0, len(zones))
	for _, z := range zones {
		data, err := ConvertVolumeTypeZone(z, projectID, region, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("convert volume type zone: %w", err)
		}
		dataList = append(dataList, data)
	}

	if err := s.saveVolumeTypeZones(ctx, dataList); err != nil {
		return nil, fmt.Errorf("save volume type zones: %w", err)
	}

	return &IngestResult{
		VolumeTypeZoneCount: len(dataList),
		CollectedAt:         collectedAt,
		DurationMillis:      time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveVolumeTypeZones(ctx context.Context, zones []*VolumeTypeZoneData) error {
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
		existing, err := tx.BronzeGreenNodeVolumeVolumeTypeZone.Query().
			Where(bronzegreennodevolumevolumetypezone.ID(data.ID)).
			First(ctx)
		if err != nil && !entvol.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing volume type zone %s: %w", data.Name, err)
		}

		diff := DiffVolumeTypeZoneData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodeVolumeVolumeTypeZone.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for volume type zone %s: %w", data.Name, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeGreenNodeVolumeVolumeTypeZone.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetRegion(data.Region).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.PoolNameJSON != nil {
				create.SetPoolNameJSON(data.PoolNameJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create volume type zone %s: %w", data.Name, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for volume type zone %s: %w", data.Name, err)
			}
		} else {
			update := tx.BronzeGreenNodeVolumeVolumeTypeZone.UpdateOneID(data.ID).
				SetName(data.Name).
				SetCollectedAt(data.CollectedAt)

			if data.PoolNameJSON != nil {
				update.SetPoolNameJSON(data.PoolNameJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update volume type zone %s: %w", data.Name, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for volume type zone %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// DeleteStaleVolumeTypeZones removes volume type zones not collected in the latest run.
func (s *Service) DeleteStaleVolumeTypeZones(ctx context.Context, projectID, region string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGreenNodeVolumeVolumeTypeZone.Query().
		Where(
			bronzegreennodevolumevolumetypezone.ProjectID(projectID),
			bronzegreennodevolumevolumetypezone.Region(region),
			bronzegreennodevolumevolumetypezone.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale volume type zones: %w", err)
	}

	for _, z := range stale {
		if err := s.history.CloseHistory(ctx, tx, z.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for volume type zone %s: %w", z.ID, err)
		}
		if err := tx.BronzeGreenNodeVolumeVolumeTypeZone.DeleteOneID(z.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete volume type zone %s: %w", z.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
