package app_inventory

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	ents1 "danny.vn/hotpot/pkg/storage/ent/s1"
	"danny.vn/hotpot/pkg/storage/ent/s1/bronzes1appinventory"
)

// Service handles SentinelOne app inventory ingestion.
type Service struct {
	client    *Client
	entClient *ents1.Client
	history   *HistoryService
}

// NewService creates a new app inventory ingestion service.
func NewService(client *Client, entClient *ents1.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of app inventory ingestion.
type IngestResult struct {
	AppCount       int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches all app inventory entries from SentinelOne using cursor pagination.
func (s *Service) Ingest(ctx context.Context, heartbeat func()) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	totalExpected, err := s.client.GetCount()
	if err != nil {
		slog.Warn("s1 app inventory: failed to get count, continuing without total", "error", err)
	}

	var allApps []*AppInventoryData
	cursor := ""
	batchNum := 0

	for {
		batchNum++
		batch, err := s.client.GetAppsBatch(cursor)
		if err != nil {
			slog.Error("s1 app inventory batch failed", "batch", batchNum, "totalSoFar", len(allApps), "error", err)
			return nil, fmt.Errorf("get app inventory batch: %w", err)
		}

		for _, app := range batch.Apps {
			allApps = append(allApps, ConvertAppInventory(app, collectedAt))
		}

		slog.Info("s1 app inventory batch fetched", "batch", batchNum, "batchItems", len(batch.Apps), "totalFetched", len(allApps), "totalExpected", totalExpected, "hasMore", batch.HasMore)

		if heartbeat != nil {
			heartbeat()
		}

		if !batch.HasMore {
			break
		}
		cursor = batch.NextCursor
	}

	if err := s.saveApps(ctx, allApps); err != nil {
		return nil, fmt.Errorf("save app inventory: %w", err)
	}

	return &IngestResult{
		AppCount:       len(allApps),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveApps(ctx context.Context, apps []*AppInventoryData) error {
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

	activeIDs := make(map[string]struct{}, len(apps))

	for _, data := range apps {
		existing, err := tx.BronzeS1AppInventory.Query().
			Where(bronzes1appinventory.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ents1.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing app inventory %s: %w", data.ResourceID, err)
		}

		diff := DiffAppInventoryData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeS1AppInventory.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for app inventory %s: %w", data.ResourceID, err)
			}
			activeIDs[data.ResourceID] = struct{}{}
			continue
		}

		if existing == nil {
			_, err := tx.BronzeS1AppInventory.Create().
				SetID(data.ResourceID).
				SetApplicationName(data.ApplicationName).
				SetApplicationVendor(data.ApplicationVendor).
				SetEndpointsCount(data.EndpointsCount).
				SetApplicationVersionsCount(data.ApplicationVersionsCount).
				SetEstimate(data.Estimate).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create app inventory %s: %w", data.ResourceID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for app inventory %s: %w", data.ResourceID, err)
			}
		} else {
			_, err := tx.BronzeS1AppInventory.UpdateOneID(data.ResourceID).
				SetApplicationName(data.ApplicationName).
				SetApplicationVendor(data.ApplicationVendor).
				SetEndpointsCount(data.EndpointsCount).
				SetApplicationVersionsCount(data.ApplicationVersionsCount).
				SetEstimate(data.Estimate).
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update app inventory %s: %w", data.ResourceID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for app inventory %s: %w", data.ResourceID, err)
			}
		}

		activeIDs[data.ResourceID] = struct{}{}
	}

	// Delete stale app inventory entries not returned by the API.
	allDBIDs, err := tx.BronzeS1AppInventory.Query().
		Select(bronzes1appinventory.FieldID).
		Strings(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query all app inventory IDs: %w", err)
	}

	staleCount := 0
	for _, id := range allDBIDs {
		if _, ok := activeIDs[id]; ok {
			continue
		}

		if err := s.history.CloseHistory(ctx, tx, id, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for stale app inventory %s: %w", id, err)
		}

		if err := tx.BronzeS1AppInventory.DeleteOneID(id).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete stale app inventory %s: %w", id, err)
		}
		staleCount++
	}

	if staleCount > 0 {
		slog.Info("s1 app inventory: deleted stale", "count", staleCount)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
