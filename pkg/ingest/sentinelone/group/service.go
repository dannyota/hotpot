package group

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	ents1 "danny.vn/hotpot/pkg/storage/ent/s1"
	"danny.vn/hotpot/pkg/storage/ent/s1/bronzes1group"
)

// Service handles SentinelOne group ingestion.
type Service struct {
	client    *Client
	entClient *ents1.Client
	history   *HistoryService
}

// NewService creates a new group ingestion service.
func NewService(client *Client, entClient *ents1.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of group ingestion.
type IngestResult struct {
	GroupCount     int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches all groups from SentinelOne using cursor pagination.
func (s *Service) Ingest(ctx context.Context, heartbeat func()) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	totalExpected, err := s.client.GetCount()
	if err != nil {
		slog.Warn("s1 groups: failed to get count, continuing without total", "error", err)
	}

	var allGroups []*GroupData
	cursor := ""
	batchNum := 0

	for {
		batchNum++
		batch, err := s.client.GetGroupsBatch(cursor)
		if err != nil {
			slog.Error("s1 groups batch failed", "batch", batchNum, "totalSoFar", len(allGroups), "error", err)
			return nil, fmt.Errorf("get groups batch: %w", err)
		}

		for _, apiGroup := range batch.Groups {
			allGroups = append(allGroups, ConvertGroup(apiGroup, collectedAt))
		}

		slog.Info("s1 groups batch fetched", "batch", batchNum, "batchItems", len(batch.Groups), "totalFetched", len(allGroups), "totalExpected", totalExpected, "hasMore", batch.HasMore)

		if heartbeat != nil {
			heartbeat()
		}

		if !batch.HasMore {
			break
		}
		cursor = batch.NextCursor
	}

	if err := s.saveGroups(ctx, allGroups); err != nil {
		return nil, fmt.Errorf("save groups: %w", err)
	}

	return &IngestResult{
		GroupCount:     len(allGroups),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveGroups(ctx context.Context, groups []*GroupData) error {
	if len(groups) == 0 {
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

	activeIDs := make(map[string]struct{}, len(groups))

	for _, data := range groups {
		existing, err := tx.BronzeS1Group.Query().
			Where(bronzes1group.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ents1.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing group %s: %w", data.ResourceID, err)
		}

		diff := DiffGroupData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeS1Group.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for group %s: %w", data.ResourceID, err)
			}
			activeIDs[data.ResourceID] = struct{}{}
			continue
		}

		if existing == nil {
			create := tx.BronzeS1Group.Create().
				SetID(data.ResourceID).
				SetName(data.Name).
				SetSiteID(data.SiteID).
				SetType(data.Type).
				SetIsDefault(data.IsDefault).
				SetInherits(data.Inherits).
				SetTotalAgents(data.TotalAgents).
				SetCreator(data.Creator).
				SetCreatorID(data.CreatorID).
				SetFilterName(data.FilterName).
				SetFilterID(data.FilterID).
				SetRegistrationToken(data.RegistrationToken).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.Rank != nil {
				create.SetRank(*data.Rank)
			}
			if data.APICreatedAt != nil {
				create.SetAPICreatedAt(*data.APICreatedAt)
			}
			if data.APIUpdatedAt != nil {
				create.SetAPIUpdatedAt(*data.APIUpdatedAt)
			}

			if _, err := create.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create group %s: %w", data.ResourceID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for group %s: %w", data.ResourceID, err)
			}
		} else {
			update := tx.BronzeS1Group.UpdateOneID(data.ResourceID).
				SetName(data.Name).
				SetSiteID(data.SiteID).
				SetType(data.Type).
				SetIsDefault(data.IsDefault).
				SetInherits(data.Inherits).
				SetTotalAgents(data.TotalAgents).
				SetCreator(data.Creator).
				SetCreatorID(data.CreatorID).
				SetFilterName(data.FilterName).
				SetFilterID(data.FilterID).
				SetRegistrationToken(data.RegistrationToken).
				SetCollectedAt(data.CollectedAt)

			if data.Rank != nil {
				update.SetRank(*data.Rank)
			} else {
				update.ClearRank()
			}
			if data.APICreatedAt != nil {
				update.SetAPICreatedAt(*data.APICreatedAt)
			}
			if data.APIUpdatedAt != nil {
				update.SetAPIUpdatedAt(*data.APIUpdatedAt)
			}

			if _, err := update.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update group %s: %w", data.ResourceID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for group %s: %w", data.ResourceID, err)
			}
		}

		activeIDs[data.ResourceID] = struct{}{}
	}

	// Delete stale groups not returned by the API.
	allDBIDs, err := tx.BronzeS1Group.Query().
		Select(bronzes1group.FieldID).
		Strings(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query all group IDs: %w", err)
	}

	staleCount := 0
	for _, id := range allDBIDs {
		if _, ok := activeIDs[id]; ok {
			continue
		}

		if err := s.history.CloseHistory(ctx, tx, id, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for stale group %s: %w", id, err)
		}

		if err := tx.BronzeS1Group.DeleteOneID(id).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete stale group %s: %w", id, err)
		}
		staleCount++
	}

	if staleCount > 0 {
		slog.Info("s1 groups: deleted stale", "count", staleCount)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
