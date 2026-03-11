package endpoint_app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	ents1 "danny.vn/hotpot/pkg/storage/ent/s1"
	"danny.vn/hotpot/pkg/storage/ent/s1/bronzes1agent"
	"danny.vn/hotpot/pkg/storage/ent/s1/bronzes1endpointapp"
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

// SaveAgentApps saves endpoint apps for a single agent (upsert + history).
func (s *Service) SaveAgentApps(ctx context.Context, agentID string, apps []*EndpointAppData) error {
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

	activeIDs := make(map[string]struct{}, len(apps))

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
			activeIDs[data.ResourceID] = struct{}{}
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

			activeIDs[data.ResourceID] = struct{}{}
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

			activeIDs[data.ResourceID] = struct{}{}
		}
	}

	// Delete stale endpoint apps for this agent not returned by the API.
	dbAppIDs, err := tx.BronzeS1EndpointApp.Query().
		Where(bronzes1endpointapp.AgentIDEQ(agentID)).
		Select(bronzes1endpointapp.FieldID).
		Strings(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query endpoint app IDs for agent %s: %w", agentID, err)
	}

	staleCount := 0
	for _, id := range dbAppIDs {
		if _, ok := activeIDs[id]; ok {
			continue
		}

		if err := s.history.CloseHistory(ctx, tx, id, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for stale endpoint app %s: %w", id, err)
		}

		if err := tx.BronzeS1EndpointApp.DeleteOneID(id).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete stale endpoint app %s: %w", id, err)
		}
		staleCount++
	}

	if staleCount > 0 {
		slog.Info("s1 endpoint apps: deleted stale for agent", "agentID", agentID, "count", staleCount)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteOrphans removes endpoint apps whose agent no longer exists.
func (s *Service) DeleteOrphans(ctx context.Context) error {
	now := time.Now()

	agentIDs, err := s.entClient.BronzeS1Agent.Query().
		Select(bronzes1agent.FieldID).
		Strings(ctx)
	if err != nil {
		return fmt.Errorf("query agent IDs: %w", err)
	}

	orphans, err := s.entClient.BronzeS1EndpointApp.Query().
		Where(bronzes1endpointapp.AgentIDNotIn(agentIDs...)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("query orphan endpoint apps: %w", err)
	}

	if len(orphans) == 0 {
		return nil
	}

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

	for _, app := range orphans {
		if err := s.history.CloseHistory(ctx, tx, app.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for orphan endpoint app %s: %w", app.ID, err)
		}

		if err := tx.BronzeS1EndpointApp.DeleteOne(app).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete orphan endpoint app %s: %w", app.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	slog.Info("s1 endpoint apps: deleted orphans", "count", len(orphans))

	return nil
}
