package quota

import (
	"context"
	"fmt"
	"time"

	entportal "github.com/dannyota/hotpot/pkg/storage/ent/greennode/portal"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/portal/bronzegreennodeportalquota"
)

// Service handles GreenNode quota ingestion.
type Service struct {
	client    *Client
	entClient *entportal.Client
	history   *HistoryService
}

// NewService creates a new quota ingestion service.
func NewService(client *Client, entClient *entportal.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of quota ingestion.
type IngestResult struct {
	QuotaCount     int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches quotas from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID, region string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	quotas, err := s.client.ListQuotas(ctx)
	if err != nil {
		return nil, fmt.Errorf("list quotas: %w", err)
	}

	dataList := make([]*QuotaData, 0, len(quotas))
	for _, q := range quotas {
		dataList = append(dataList, ConvertQuota(q, projectID, region, collectedAt))
	}

	if err := s.saveQuotas(ctx, dataList); err != nil {
		return nil, fmt.Errorf("save quotas: %w", err)
	}

	return &IngestResult{
		QuotaCount:     len(dataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveQuotas(ctx context.Context, quotas []*QuotaData) error {
	if len(quotas) == 0 {
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

	for _, data := range quotas {
		existing, err := tx.BronzeGreenNodePortalQuota.Query().
			Where(bronzegreennodeportalquota.ID(data.ID)).
			First(ctx)
		if err != nil && !entportal.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing quota %s: %w", data.Name, err)
		}

		diff := DiffQuotaData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodePortalQuota.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for quota %s: %w", data.Name, err)
			}
			continue
		}

		if existing == nil {
			_, err = tx.BronzeGreenNodePortalQuota.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetType(data.Type).
				SetLimitValue(data.LimitValue).
				SetUsedValue(data.UsedValue).
				SetRegion(data.Region).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create quota %s: %w", data.Name, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for quota %s: %w", data.Name, err)
			}
		} else {
			_, err = tx.BronzeGreenNodePortalQuota.UpdateOneID(data.ID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetType(data.Type).
				SetLimitValue(data.LimitValue).
				SetUsedValue(data.UsedValue).
				SetRegion(data.Region).
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update quota %s: %w", data.Name, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for quota %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// DeleteStaleQuotas removes quotas not collected in the latest run for the given region.
func (s *Service) DeleteStaleQuotas(ctx context.Context, projectID, region string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGreenNodePortalQuota.Query().
		Where(
			bronzegreennodeportalquota.ProjectID(projectID),
			bronzegreennodeportalquota.Region(region),
			bronzegreennodeportalquota.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale quotas: %w", err)
	}

	for _, q := range stale {
		if err := s.history.CloseHistory(ctx, tx, q.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for quota %s: %w", q.ID, err)
		}
		if err := tx.BronzeGreenNodePortalQuota.DeleteOneID(q.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete quota %s: %w", q.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
