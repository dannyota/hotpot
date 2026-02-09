package group

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzes1group"
)

// Service handles SentinelOne group ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new group ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
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

	var allGroups []*GroupData
	cursor := ""

	for {
		batch, err := s.client.GetGroupsBatch(cursor)
		if err != nil {
			return nil, fmt.Errorf("get groups batch: %w", err)
		}

		for _, apiGroup := range batch.Groups {
			allGroups = append(allGroups, ConvertGroup(apiGroup, collectedAt))
		}

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

	for _, data := range groups {
		existing, err := tx.BronzeS1Group.Query().
			Where(bronzes1group.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
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
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStale removes groups that were not collected in the latest run.
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

	stale, err := tx.BronzeS1Group.Query().
		Where(bronzes1group.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, g := range stale {
		if err := s.history.CloseHistory(ctx, tx, g.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for group %s: %w", g.ID, err)
		}

		if err := tx.BronzeS1Group.DeleteOne(g).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete group %s: %w", g.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
