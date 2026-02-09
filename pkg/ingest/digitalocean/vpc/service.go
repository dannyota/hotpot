package vpc

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzedovpc"
)

// Service handles DigitalOcean VPC ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new VPC ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of VPC ingestion.
type IngestResult struct {
	VpcCount       int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches all VPCs from DigitalOcean and saves them.
func (s *Service) Ingest(ctx context.Context, heartbeat func()) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	apiVPCs, err := s.client.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list VPCs: %w", err)
	}

	if heartbeat != nil {
		heartbeat()
	}

	var allVPCs []*VpcData
	for _, v := range apiVPCs {
		allVPCs = append(allVPCs, ConvertVpc(v, collectedAt))
	}

	if err := s.saveVPCs(ctx, allVPCs); err != nil {
		return nil, fmt.Errorf("save VPCs: %w", err)
	}

	return &IngestResult{
		VpcCount:       len(allVPCs),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveVPCs(ctx context.Context, vpcs []*VpcData) error {
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
		existing, err := tx.BronzeDOVpc.Query().
			Where(bronzedovpc.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing VPC %s: %w", data.ResourceID, err)
		}

		diff := DiffVpcData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeDOVpc.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for VPC %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeDOVpc.Create().
				SetID(data.ResourceID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetRegion(data.Region).
				SetIPRange(data.IPRange).
				SetUrn(data.URN).
				SetIsDefault(data.IsDefault).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.APICreatedAt != nil {
				create.SetAPICreatedAt(*data.APICreatedAt)
			}

			if _, err := create.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create VPC %s: %w", data.ResourceID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for VPC %s: %w", data.ResourceID, err)
			}
		} else {
			update := tx.BronzeDOVpc.UpdateOneID(data.ResourceID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetRegion(data.Region).
				SetIPRange(data.IPRange).
				SetUrn(data.URN).
				SetIsDefault(data.IsDefault).
				SetCollectedAt(data.CollectedAt)

			if data.APICreatedAt != nil {
				update.SetAPICreatedAt(*data.APICreatedAt)
			}

			if _, err := update.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update VPC %s: %w", data.ResourceID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for VPC %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStale removes VPCs that were not collected in the latest run.
func (s *Service) DeleteStale(ctx context.Context, collectedAt time.Time) error {
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

	stale, err := tx.BronzeDOVpc.Query().
		Where(bronzedovpc.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, doVpc := range stale {
		if err := s.history.CloseHistory(ctx, tx, doVpc.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for VPC %s: %w", doVpc.ID, err)
		}

		if err := tx.BronzeDOVpc.DeleteOne(doVpc).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete VPC %s: %w", doVpc.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
