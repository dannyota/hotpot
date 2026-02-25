package interconnect

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegreennodenetworkinterconnect"
)

// Service handles GreenNode interconnect ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new interconnect ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of interconnect ingestion.
type IngestResult struct {
	InterconnectCount int
	CollectedAt       time.Time
	DurationMillis    int64
}

// Ingest fetches interconnects from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID, region string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	interconnects, err := s.client.ListInterconnects(ctx)
	if err != nil {
		return nil, fmt.Errorf("list interconnects: %w", err)
	}

	dataList := make([]*InterconnectData, 0, len(interconnects))
	for _, ic := range interconnects {
		dataList = append(dataList, ConvertInterconnect(ic, projectID, region, collectedAt))
	}

	if err := s.saveInterconnects(ctx, dataList); err != nil {
		return nil, fmt.Errorf("save interconnects: %w", err)
	}

	return &IngestResult{
		InterconnectCount: len(dataList),
		CollectedAt:       collectedAt,
		DurationMillis:    time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveInterconnects(ctx context.Context, interconnects []*InterconnectData) error {
	if len(interconnects) == 0 {
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

	for _, data := range interconnects {
		existing, err := tx.BronzeGreenNodeNetworkInterconnect.Query().
			Where(bronzegreennodenetworkinterconnect.ID(data.UUID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing interconnect %s: %w", data.Name, err)
		}

		diff := DiffInterconnectData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodeNetworkInterconnect.UpdateOneID(data.UUID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for interconnect %s: %w", data.Name, err)
			}
			continue
		}

		if existing == nil {
			_, err = tx.BronzeGreenNodeNetworkInterconnect.Create().
				SetID(data.UUID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetStatus(data.Status).
				SetEnableGw2(data.EnableGw2).
				SetCircuitID(data.CircuitID).
				SetGw01IP(data.Gw01IP).
				SetGw02IP(data.Gw02IP).
				SetGwVip(data.GwVIP).
				SetRemoteGw01IP(data.RemoteGw01IP).
				SetRemoteGw02IP(data.RemoteGw02IP).
				SetPackageID(data.PackageID).
				SetTypeID(data.TypeID).
				SetTypeName(data.TypeName).
				SetCreatedAt(data.CreatedAt).
				SetRegion(data.Region).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create interconnect %s: %w", data.Name, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for interconnect %s: %w", data.Name, err)
			}
		} else {
			_, err = tx.BronzeGreenNodeNetworkInterconnect.UpdateOneID(data.UUID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetStatus(data.Status).
				SetEnableGw2(data.EnableGw2).
				SetCircuitID(data.CircuitID).
				SetGw01IP(data.Gw01IP).
				SetGw02IP(data.Gw02IP).
				SetGwVip(data.GwVIP).
				SetRemoteGw01IP(data.RemoteGw01IP).
				SetRemoteGw02IP(data.RemoteGw02IP).
				SetPackageID(data.PackageID).
				SetTypeID(data.TypeID).
				SetTypeName(data.TypeName).
				SetCreatedAt(data.CreatedAt).
				SetRegion(data.Region).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update interconnect %s: %w", data.Name, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for interconnect %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// DeleteStaleInterconnects removes interconnects not collected in the latest run for the given region.
func (s *Service) DeleteStaleInterconnects(ctx context.Context, projectID, region string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGreenNodeNetworkInterconnect.Query().
		Where(
			bronzegreennodenetworkinterconnect.ProjectID(projectID),
			bronzegreennodenetworkinterconnect.Region(region),
			bronzegreennodenetworkinterconnect.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale interconnects: %w", err)
	}

	for _, ic := range stale {
		if err := s.history.CloseHistory(ctx, tx, ic.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for interconnect %s: %w", ic.ID, err)
		}
		if err := tx.BronzeGreenNodeNetworkInterconnect.DeleteOneID(ic.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete interconnect %s: %w", ic.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
