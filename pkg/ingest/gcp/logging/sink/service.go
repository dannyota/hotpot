package sink

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcploggingsink"
)

// Service handles GCP Cloud Logging sink ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new sink ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for sink ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of sink ingestion.
type IngestResult struct {
	ProjectID      string
	SinkCount      int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches sinks from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	sinks, err := s.client.ListSinks(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sinks: %w", err)
	}

	sinkDataList := make([]*SinkData, 0, len(sinks))
	for _, s := range sinks {
		data, err := ConvertSink(s, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert sink: %w", err)
		}
		sinkDataList = append(sinkDataList, data)
	}

	if err := s.saveSinks(ctx, sinkDataList); err != nil {
		return nil, fmt.Errorf("failed to save sinks: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		SinkCount:      len(sinkDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveSinks saves sinks to the database with history tracking.
func (s *Service) saveSinks(ctx context.Context, sinks []*SinkData) error {
	if len(sinks) == 0 {
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

	for _, sinkData := range sinks {
		existing, err := tx.BronzeGCPLoggingSink.Query().
			Where(bronzegcploggingsink.ID(sinkData.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing sink %s: %w", sinkData.Name, err)
		}

		diff := DiffSinkData(existing, sinkData)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPLoggingSink.UpdateOneID(sinkData.ResourceID).
				SetCollectedAt(sinkData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for sink %s: %w", sinkData.Name, err)
			}
			continue
		}

		if existing == nil {
			// Create new sink
			create := tx.BronzeGCPLoggingSink.Create().
				SetID(sinkData.ResourceID).
				SetName(sinkData.Name).
				SetDestination(sinkData.Destination).
				SetFilter(sinkData.Filter).
				SetDescription(sinkData.Description).
				SetDisabled(sinkData.Disabled).
				SetIncludeChildren(sinkData.IncludeChildren).
				SetWriterIdentity(sinkData.WriterIdentity).
				SetProjectID(sinkData.ProjectID).
				SetCollectedAt(sinkData.CollectedAt).
				SetFirstCollectedAt(sinkData.CollectedAt)

			if sinkData.ExclusionsJSON != nil {
				create.SetExclusionsJSON(sinkData.ExclusionsJSON)
			}
			if sinkData.BigqueryOptionsJSON != nil {
				create.SetBigqueryOptionsJSON(sinkData.BigqueryOptionsJSON)
			}

			_, err := create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create sink %s: %w", sinkData.Name, err)
			}
		} else {
			// Update existing sink
			update := tx.BronzeGCPLoggingSink.UpdateOneID(sinkData.ResourceID).
				SetName(sinkData.Name).
				SetDestination(sinkData.Destination).
				SetFilter(sinkData.Filter).
				SetDescription(sinkData.Description).
				SetDisabled(sinkData.Disabled).
				SetIncludeChildren(sinkData.IncludeChildren).
				SetWriterIdentity(sinkData.WriterIdentity).
				SetProjectID(sinkData.ProjectID).
				SetCollectedAt(sinkData.CollectedAt)

			if sinkData.ExclusionsJSON != nil {
				update.SetExclusionsJSON(sinkData.ExclusionsJSON)
			}
			if sinkData.BigqueryOptionsJSON != nil {
				update.SetBigqueryOptionsJSON(sinkData.BigqueryOptionsJSON)
			}

			_, err := update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update sink %s: %w", sinkData.Name, err)
			}
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, sinkData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for sink %s: %w", sinkData.Name, err)
			}
		} else if diff.IsChanged {
			if err := s.history.UpdateHistory(ctx, tx, existing, sinkData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for sink %s: %w", sinkData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleSinks removes sinks that were not collected in the latest run.
func (s *Service) DeleteStaleSinks(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGCPLoggingSink.Query().
		Where(
			bronzegcploggingsink.ProjectID(projectID),
			bronzegcploggingsink.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, staleSink := range stale {
		if err := s.history.CloseHistory(ctx, tx, staleSink.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for sink %s: %w", staleSink.ID, err)
		}
		if err := tx.BronzeGCPLoggingSink.DeleteOne(staleSink).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete sink %s: %w", staleSink.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
