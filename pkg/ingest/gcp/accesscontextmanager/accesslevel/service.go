package accesslevel

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpaccesscontextmanageraccesslevel"
)

// Service handles access level ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new access level ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of access level ingestion.
type IngestResult struct {
	LevelCount     int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches access levels from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch access levels from GCP
	rawLevels, err := s.client.ListAccessLevels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list access levels: %w", err)
	}

	// Convert to level data
	levelDataList := make([]*AccessLevelData, 0, len(rawLevels))
	for _, raw := range rawLevels {
		data := ConvertAccessLevel(raw.OrgName, raw.AccessPolicyName, raw.AccessLevel, collectedAt)
		levelDataList = append(levelDataList, data)
	}

	// Save to database
	if err := s.saveLevels(ctx, levelDataList); err != nil {
		return nil, fmt.Errorf("failed to save access levels: %w", err)
	}

	return &IngestResult{
		LevelCount:     len(levelDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveLevels saves access levels to the database with history tracking.
func (s *Service) saveLevels(ctx context.Context, levels []*AccessLevelData) error {
	if len(levels) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, levelData := range levels {
		// Load existing level
		existing, err := tx.BronzeGCPAccessContextManagerAccessLevel.Query().
			Where(bronzegcpaccesscontextmanageraccesslevel.ID(levelData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing access level %s: %w", levelData.ID, err)
		}

		// Compute diff
		diff := DiffAccessLevelData(existing, levelData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPAccessContextManagerAccessLevel.UpdateOneID(levelData.ID).
				SetCollectedAt(levelData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for access level %s: %w", levelData.ID, err)
			}
			continue
		}

		// Create or update level
		if existing == nil {
			create := tx.BronzeGCPAccessContextManagerAccessLevel.Create().
				SetID(levelData.ID).
				SetAccessPolicyName(levelData.AccessPolicyName).
				SetAccessPolicyID(levelData.AccessPolicyName).
				SetOrganizationID(levelData.OrganizationID).
				SetCollectedAt(levelData.CollectedAt).
				SetFirstCollectedAt(levelData.CollectedAt)

			if levelData.Title != "" {
				create.SetTitle(levelData.Title)
			}
			if levelData.Description != "" {
				create.SetDescription(levelData.Description)
			}
			if levelData.BasicJSON != nil {
				create.SetBasicJSON(levelData.BasicJSON)
			}
			if levelData.CustomJSON != nil {
				create.SetCustomJSON(levelData.CustomJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create access level %s: %w", levelData.ID, err)
			}
		} else {
			update := tx.BronzeGCPAccessContextManagerAccessLevel.UpdateOneID(levelData.ID).
				SetAccessPolicyName(levelData.AccessPolicyName).
				SetAccessPolicyID(levelData.AccessPolicyName).
				SetOrganizationID(levelData.OrganizationID).
				SetCollectedAt(levelData.CollectedAt)

			if levelData.Title != "" {
				update.SetTitle(levelData.Title)
			}
			if levelData.Description != "" {
				update.SetDescription(levelData.Description)
			}
			if levelData.BasicJSON != nil {
				update.SetBasicJSON(levelData.BasicJSON)
			}
			if levelData.CustomJSON != nil {
				update.SetCustomJSON(levelData.CustomJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update access level %s: %w", levelData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, levelData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for access level %s: %w", levelData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, levelData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for access level %s: %w", levelData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleLevels removes access levels that were not collected in the latest run.
func (s *Service) DeleteStaleLevels(ctx context.Context, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	staleLevels, err := tx.BronzeGCPAccessContextManagerAccessLevel.Query().
		Where(bronzegcpaccesscontextmanageraccesslevel.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, level := range staleLevels {
		if err := s.history.CloseHistory(ctx, tx, level.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for access level %s: %w", level.ID, err)
		}

		if err := tx.BronzeGCPAccessContextManagerAccessLevel.DeleteOne(level).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete access level %s: %w", level.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
