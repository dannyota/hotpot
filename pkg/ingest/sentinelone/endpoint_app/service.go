package endpoint_app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	ents1 "github.com/dannyota/hotpot/pkg/storage/ent/s1"
	"github.com/dannyota/hotpot/pkg/storage/ent/s1/bronzes1agent"
	"github.com/dannyota/hotpot/pkg/storage/ent/s1/bronzes1endpointapp"
)

// Service handles SentinelOne endpoint app ingestion.
type Service struct {
	client    *Client
	entClient *ents1.Client
	history   *HistoryService
}

// NewService creates a new endpoint app ingestion service.
func NewService(client *Client, entClient *ents1.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of endpoint app ingestion.
type IngestResult struct {
	AppCount       int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches endpoint apps by querying agent IDs from the database first.
func (s *Service) Ingest(ctx context.Context, heartbeat func()) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Step 1: Get all agent IDs from database
	agents, err := s.entClient.BronzeS1Agent.Query().
		Select(bronzes1agent.FieldID).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query agent IDs: %w", err)
	}

	slog.Info("s1 endpoint apps: fetched agent IDs", "agentCount", len(agents))

	// Step 2: For each agent, fetch their apps
	var allApps []*EndpointAppData
	for i, agent := range agents {
		apiApps, err := s.client.GetEndpointApps(agent.ID)
		if err != nil {
			return nil, fmt.Errorf("get endpoint apps for agent %s: %w", agent.ID, err)
		}

		for _, app := range apiApps {
			allApps = append(allApps, ConvertEndpointApp(agent.ID, app, collectedAt))
		}

		if (i+1)%50 == 0 {
			slog.Info("s1 endpoint apps: progress", "agentsProcessed", i+1, "totalAgents", len(agents), "totalApps", len(allApps))
		}

		if heartbeat != nil {
			heartbeat()
		}
	}

	slog.Info("s1 endpoint apps: fetch complete", "totalAgents", len(agents), "totalApps", len(allApps))

	// Step 3: Save + history + delete stale
	if err := s.saveApps(ctx, allApps); err != nil {
		return nil, fmt.Errorf("save endpoint apps: %w", err)
	}

	return &IngestResult{
		AppCount:       len(allApps),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveApps(ctx context.Context, apps []*EndpointAppData) error {
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
		existing, err := tx.BronzeS1EndpointApp.Query().
			Where(bronzes1endpointapp.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ents1.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing endpoint app %s: %w", data.ResourceID, err)
		}

		diff := DiffEndpointAppData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeS1EndpointApp.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for endpoint app %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeS1EndpointApp.Create().
				SetID(data.ResourceID).
				SetAgentID(data.AgentID).
				SetName(data.Name).
				SetVersion(data.Version).
				SetPublisher(data.Publisher).
				SetSize(data.Size).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.InstalledDate != nil {
				create.SetInstalledDate(*data.InstalledDate)
			}

			if _, err := create.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create endpoint app %s: %w", data.ResourceID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for endpoint app %s: %w", data.ResourceID, err)
			}
		} else {
			update := tx.BronzeS1EndpointApp.UpdateOneID(data.ResourceID).
				SetAgentID(data.AgentID).
				SetName(data.Name).
				SetVersion(data.Version).
				SetPublisher(data.Publisher).
				SetSize(data.Size).
				SetCollectedAt(data.CollectedAt)

			if data.InstalledDate != nil {
				update.SetInstalledDate(*data.InstalledDate)
			}

			if _, err := update.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update endpoint app %s: %w", data.ResourceID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for endpoint app %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStale removes endpoint apps that were not collected in the latest run.
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

	stale, err := tx.BronzeS1EndpointApp.Query().
		Where(bronzes1endpointapp.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, app := range stale {
		if err := s.history.CloseHistory(ctx, tx, app.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for endpoint app %s: %w", app.ID, err)
		}

		if err := tx.BronzeS1EndpointApp.DeleteOne(app).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete endpoint app %s: %w", app.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
