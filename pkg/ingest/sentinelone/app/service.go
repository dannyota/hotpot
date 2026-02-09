package app

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzes1app"
)

// Service handles SentinelOne installed application ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new app ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of app ingestion.
type IngestResult struct {
	AppCount       int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches all installed applications from SentinelOne using cursor pagination.
func (s *Service) Ingest(ctx context.Context, heartbeat func()) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	var allApps []*AppData
	cursor := ""

	for {
		batch, err := s.client.GetAppsBatch(cursor)
		if err != nil {
			return nil, fmt.Errorf("get apps batch: %w", err)
		}

		for _, apiApp := range batch.Apps {
			allApps = append(allApps, ConvertApp(apiApp, collectedAt))
		}

		if heartbeat != nil {
			heartbeat()
		}

		if !batch.HasMore {
			break
		}
		cursor = batch.NextCursor
	}

	if err := s.saveApps(ctx, allApps); err != nil {
		return nil, fmt.Errorf("save apps: %w", err)
	}

	return &IngestResult{
		AppCount:       len(allApps),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveApps(ctx context.Context, apps []*AppData) error {
	if len(apps) == 0 {
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

	for _, data := range apps {
		existing, err := tx.BronzeS1App.Query().
			Where(bronzes1app.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing app %s: %w", data.ResourceID, err)
		}

		diff := DiffAppData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeS1App.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for app %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeS1App.Create().
				SetID(data.ResourceID).
				SetName(data.Name).
				SetPublisher(data.Publisher).
				SetVersion(data.Version).
				SetSize(data.Size).
				SetAppType(data.AppType).
				SetOsType(data.OsType).
				SetAgentID(data.AgentID).
				SetAgentComputerName(data.AgentComputerName).
				SetAgentMachineType(data.AgentMachineType).
				SetAgentIsActive(data.AgentIsActive).
				SetAgentIsDecommissioned(data.AgentIsDecommissioned).
				SetRiskLevel(data.RiskLevel).
				SetSigned(data.Signed).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.InstalledDate != nil {
				create.SetInstalledDate(*data.InstalledDate)
			}

			if _, err := create.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create app %s: %w", data.ResourceID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for app %s: %w", data.ResourceID, err)
			}
		} else {
			update := tx.BronzeS1App.UpdateOneID(data.ResourceID).
				SetName(data.Name).
				SetPublisher(data.Publisher).
				SetVersion(data.Version).
				SetSize(data.Size).
				SetAppType(data.AppType).
				SetOsType(data.OsType).
				SetAgentID(data.AgentID).
				SetAgentComputerName(data.AgentComputerName).
				SetAgentMachineType(data.AgentMachineType).
				SetAgentIsActive(data.AgentIsActive).
				SetAgentIsDecommissioned(data.AgentIsDecommissioned).
				SetRiskLevel(data.RiskLevel).
				SetSigned(data.Signed).
				SetCollectedAt(data.CollectedAt)

			if data.InstalledDate != nil {
				update.SetInstalledDate(*data.InstalledDate)
			}

			if _, err := update.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update app %s: %w", data.ResourceID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for app %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStale removes apps that were not collected in the latest run.
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

	stale, err := tx.BronzeS1App.Query().
		Where(bronzes1app.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, a := range stale {
		if err := s.history.CloseHistory(ctx, tx, a.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for app %s: %w", a.ID, err)
		}

		if err := tx.BronzeS1App.DeleteOne(a).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete app %s: %w", a.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
