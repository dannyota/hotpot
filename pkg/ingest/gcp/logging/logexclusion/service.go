package logexclusion

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcplogginglogexclusion"
)

// Service handles GCP Cloud Logging log exclusion ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new log exclusion ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for log exclusion ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of log exclusion ingestion.
type IngestResult struct {
	ProjectID      string
	ExclusionCount int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches log exclusions from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	exclusions, err := s.client.ListExclusions(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list exclusions: %w", err)
	}

	exclusionDataList := make([]*LogExclusionData, 0, len(exclusions))
	for _, e := range exclusions {
		data := ConvertExclusion(e, params.ProjectID, collectedAt)
		exclusionDataList = append(exclusionDataList, data)
	}

	if err := s.saveExclusions(ctx, exclusionDataList); err != nil {
		return nil, fmt.Errorf("failed to save exclusions: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		ExclusionCount: len(exclusionDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveExclusions saves log exclusions to the database with history tracking.
func (s *Service) saveExclusions(ctx context.Context, exclusions []*LogExclusionData) error {
	if len(exclusions) == 0 {
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

	for _, exclusionData := range exclusions {
		existing, err := tx.BronzeGCPLoggingLogExclusion.Query().
			Where(bronzegcplogginglogexclusion.ID(exclusionData.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing exclusion %s: %w", exclusionData.Name, err)
		}

		diff := DiffExclusionData(existing, exclusionData)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPLoggingLogExclusion.UpdateOneID(exclusionData.ResourceID).
				SetCollectedAt(exclusionData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for exclusion %s: %w", exclusionData.Name, err)
			}
			continue
		}

		if existing == nil {
			// Create new exclusion
			_, err := tx.BronzeGCPLoggingLogExclusion.Create().
				SetID(exclusionData.ResourceID).
				SetName(exclusionData.Name).
				SetDescription(exclusionData.Description).
				SetFilter(exclusionData.Filter).
				SetDisabled(exclusionData.Disabled).
				SetCreateTime(exclusionData.CreateTime).
				SetUpdateTime(exclusionData.UpdateTime).
				SetProjectID(exclusionData.ProjectID).
				SetCollectedAt(exclusionData.CollectedAt).
				SetFirstCollectedAt(exclusionData.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create exclusion %s: %w", exclusionData.Name, err)
			}
		} else {
			// Update existing exclusion
			_, err := tx.BronzeGCPLoggingLogExclusion.UpdateOneID(exclusionData.ResourceID).
				SetName(exclusionData.Name).
				SetDescription(exclusionData.Description).
				SetFilter(exclusionData.Filter).
				SetDisabled(exclusionData.Disabled).
				SetCreateTime(exclusionData.CreateTime).
				SetUpdateTime(exclusionData.UpdateTime).
				SetProjectID(exclusionData.ProjectID).
				SetCollectedAt(exclusionData.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update exclusion %s: %w", exclusionData.Name, err)
			}
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, exclusionData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for exclusion %s: %w", exclusionData.Name, err)
			}
		} else if diff.IsChanged {
			if err := s.history.UpdateHistory(ctx, tx, existing, exclusionData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for exclusion %s: %w", exclusionData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleExclusions removes log exclusions that were not collected in the latest run.
func (s *Service) DeleteStaleExclusions(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGCPLoggingLogExclusion.Query().
		Where(
			bronzegcplogginglogexclusion.ProjectID(projectID),
			bronzegcplogginglogexclusion.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, staleExclusion := range stale {
		if err := s.history.CloseHistory(ctx, tx, staleExclusion.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for exclusion %s: %w", staleExclusion.ID, err)
		}
		if err := tx.BronzeGCPLoggingLogExclusion.DeleteOne(staleExclusion).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete exclusion %s: %w", staleExclusion.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
