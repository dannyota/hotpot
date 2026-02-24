package volumetype

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegreennodevolumevolumetype"
)

// Service handles GreenNode volume type ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new volume type ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of volume type ingestion.
type IngestResult struct {
	VolumeTypeCount int
	CollectedAt     time.Time
	DurationMillis  int64
}

// Ingest fetches volume types from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID, region string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	volumeTypes, err := s.client.ListVolumeTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("list volume types: %w", err)
	}

	dataList := make([]*VolumeTypeData, 0, len(volumeTypes))
	for _, vt := range volumeTypes {
		dataList = append(dataList, ConvertVolumeType(vt, projectID, region, collectedAt))
	}

	if err := s.saveVolumeTypes(ctx, dataList); err != nil {
		return nil, fmt.Errorf("save volume types: %w", err)
	}

	return &IngestResult{
		VolumeTypeCount: len(dataList),
		CollectedAt:     collectedAt,
		DurationMillis:  time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveVolumeTypes(ctx context.Context, volumeTypes []*VolumeTypeData) error {
	if len(volumeTypes) == 0 {
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

	for _, data := range volumeTypes {
		existing, err := tx.BronzeGreenNodeVolumeVolumeType.Query().
			Where(bronzegreennodevolumevolumetype.ID(data.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing volume type %s: %w", data.Name, err)
		}

		diff := DiffVolumeTypeData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodeVolumeVolumeType.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for volume type %s: %w", data.Name, err)
			}
			continue
		}

		if existing == nil {
			_, err = tx.BronzeGreenNodeVolumeVolumeType.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetIops(data.Iops).
				SetMaxSize(data.MaxSize).
				SetMinSize(data.MinSize).
				SetThroughPut(data.ThroughPut).
				SetZoneID(data.ZoneID).
				SetRegion(data.Region).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create volume type %s: %w", data.Name, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for volume type %s: %w", data.Name, err)
			}
		} else {
			_, err = tx.BronzeGreenNodeVolumeVolumeType.UpdateOneID(data.ID).
				SetName(data.Name).
				SetIops(data.Iops).
				SetMaxSize(data.MaxSize).
				SetMinSize(data.MinSize).
				SetThroughPut(data.ThroughPut).
				SetZoneID(data.ZoneID).
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update volume type %s: %w", data.Name, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for volume type %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// DeleteStaleVolumeTypes removes volume types not collected in the latest run.
func (s *Service) DeleteStaleVolumeTypes(ctx context.Context, projectID, region string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGreenNodeVolumeVolumeType.Query().
		Where(
			bronzegreennodevolumevolumetype.ProjectID(projectID),
			bronzegreennodevolumevolumetype.Region(region),
			bronzegreennodevolumevolumetype.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale volume types: %w", err)
	}

	for _, vt := range stale {
		if err := s.history.CloseHistory(ctx, tx, vt.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for volume type %s: %w", vt.ID, err)
		}
		if err := tx.BronzeGreenNodeVolumeVolumeType.DeleteOneID(vt.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete volume type %s: %w", vt.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
