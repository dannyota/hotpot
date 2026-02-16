package source

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpsecuritycentersource"
)

// Service handles SCC source ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new SCC source ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of SCC source ingestion.
type IngestResult struct {
	SourceCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches SCC sources from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch sources from GCP
	rawSources, err := s.client.ListSources(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list sources: %w", err)
	}

	// Convert to source data
	sourceDataList := make([]*SourceData, 0, len(rawSources))
	for _, raw := range rawSources {
		data := ConvertSource(raw.OrgName, raw.Source, collectedAt)
		sourceDataList = append(sourceDataList, data)
	}

	// Save to database
	if err := s.saveSources(ctx, sourceDataList); err != nil {
		return nil, fmt.Errorf("failed to save sources: %w", err)
	}

	return &IngestResult{
		SourceCount:    len(sourceDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveSources saves SCC sources to the database with history tracking.
func (s *Service) saveSources(ctx context.Context, sources []*SourceData) error {
	if len(sources) == 0 {
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

	for _, sourceData := range sources {
		// Load existing source
		existing, err := tx.BronzeGCPSecurityCenterSource.Query().
			Where(bronzegcpsecuritycentersource.ID(sourceData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing source %s: %w", sourceData.ID, err)
		}

		// Compute diff
		diff := DiffSourceData(existing, sourceData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPSecurityCenterSource.UpdateOneID(sourceData.ID).
				SetCollectedAt(sourceData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for source %s: %w", sourceData.ID, err)
			}
			continue
		}

		// Create or update source
		if existing == nil {
			create := tx.BronzeGCPSecurityCenterSource.Create().
				SetID(sourceData.ID).
				SetOrganizationID(sourceData.OrganizationID).
				SetCollectedAt(sourceData.CollectedAt).
				SetFirstCollectedAt(sourceData.CollectedAt)

			if sourceData.DisplayName != "" {
				create.SetDisplayName(sourceData.DisplayName)
			}
			if sourceData.Description != "" {
				create.SetDescription(sourceData.Description)
			}
			if sourceData.CanonicalName != "" {
				create.SetCanonicalName(sourceData.CanonicalName)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create source %s: %w", sourceData.ID, err)
			}
		} else {
			update := tx.BronzeGCPSecurityCenterSource.UpdateOneID(sourceData.ID).
				SetOrganizationID(sourceData.OrganizationID).
				SetCollectedAt(sourceData.CollectedAt)

			if sourceData.DisplayName != "" {
				update.SetDisplayName(sourceData.DisplayName)
			}
			if sourceData.Description != "" {
				update.SetDescription(sourceData.Description)
			}
			if sourceData.CanonicalName != "" {
				update.SetCanonicalName(sourceData.CanonicalName)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update source %s: %w", sourceData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, sourceData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for source %s: %w", sourceData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, sourceData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for source %s: %w", sourceData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleSources removes sources that were not collected in the latest run.
func (s *Service) DeleteStaleSources(ctx context.Context, collectedAt time.Time) error {
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

	staleSources, err := tx.BronzeGCPSecurityCenterSource.Query().
		Where(bronzegcpsecuritycentersource.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, source := range staleSources {
		if err := s.history.CloseHistory(ctx, tx, source.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for source %s: %w", source.ID, err)
		}

		if err := tx.BronzeGCPSecurityCenterSource.DeleteOne(source).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete source %s: %w", source.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
