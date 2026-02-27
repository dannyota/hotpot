package endpoint_app

import (
	"context"
	"fmt"
	"time"

	ents1 "github.com/dannyota/hotpot/pkg/storage/ent/s1"
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

// SaveAgentApps saves endpoint apps for a single agent (upsert + history).
func (s *Service) SaveAgentApps(ctx context.Context, apps []*EndpointAppData) error {
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
