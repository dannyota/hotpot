package attestor

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpbinaryauthorizationattestor"
)

// Service handles Binary Authorization attestor ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new Binary Authorization attestor ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for attestor ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of Binary Authorization attestor ingestion.
type IngestResult struct {
	ProjectID      string
	AttestorCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches attestors from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	rawAttestors, err := s.client.ListAttestors(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list attestors: %w", err)
	}

	attestorDataList := make([]*AttestorData, 0, len(rawAttestors))
	for _, raw := range rawAttestors {
		data := ConvertAttestor(raw, params.ProjectID, collectedAt)
		attestorDataList = append(attestorDataList, data)
	}

	if err := s.saveAttestors(ctx, attestorDataList); err != nil {
		return nil, fmt.Errorf("failed to save attestors: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		AttestorCount:  len(attestorDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveAttestors saves Binary Authorization attestors to the database with history tracking.
func (s *Service) saveAttestors(ctx context.Context, attestors []*AttestorData) error {
	if len(attestors) == 0 {
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

	for _, attestorData := range attestors {
		// Load existing attestor
		existing, err := tx.BronzeGCPBinaryAuthorizationAttestor.Query().
			Where(bronzegcpbinaryauthorizationattestor.ID(attestorData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing attestor %s: %w", attestorData.ID, err)
		}

		// Compute diff
		diff := DiffAttestorData(existing, attestorData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPBinaryAuthorizationAttestor.UpdateOneID(attestorData.ID).
				SetCollectedAt(attestorData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for attestor %s: %w", attestorData.ID, err)
			}
			continue
		}

		// Create or update attestor
		if existing == nil {
			create := tx.BronzeGCPBinaryAuthorizationAttestor.Create().
				SetID(attestorData.ID).
				SetProjectID(attestorData.ProjectID).
				SetCollectedAt(attestorData.CollectedAt).
				SetFirstCollectedAt(attestorData.CollectedAt)

			if attestorData.Description != "" {
				create.SetDescription(attestorData.Description)
			}
			if attestorData.UserOwnedGrafeasNoteJSON != nil {
				create.SetUserOwnedGrafeasNoteJSON(attestorData.UserOwnedGrafeasNoteJSON)
			}
			if attestorData.UpdateTime != "" {
				create.SetUpdateTime(attestorData.UpdateTime)
			}
			if attestorData.Etag != "" {
				create.SetEtag(attestorData.Etag)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create attestor %s: %w", attestorData.ID, err)
			}
		} else {
			update := tx.BronzeGCPBinaryAuthorizationAttestor.UpdateOneID(attestorData.ID).
				SetProjectID(attestorData.ProjectID).
				SetCollectedAt(attestorData.CollectedAt)

			if attestorData.Description != "" {
				update.SetDescription(attestorData.Description)
			}
			if attestorData.UserOwnedGrafeasNoteJSON != nil {
				update.SetUserOwnedGrafeasNoteJSON(attestorData.UserOwnedGrafeasNoteJSON)
			}
			if attestorData.UpdateTime != "" {
				update.SetUpdateTime(attestorData.UpdateTime)
			}
			if attestorData.Etag != "" {
				update.SetEtag(attestorData.Etag)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update attestor %s: %w", attestorData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, attestorData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for attestor %s: %w", attestorData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, attestorData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for attestor %s: %w", attestorData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleAttestors removes attestors that were not collected in the latest run.
func (s *Service) DeleteStaleAttestors(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	staleAttestors, err := tx.BronzeGCPBinaryAuthorizationAttestor.Query().
		Where(
			bronzegcpbinaryauthorizationattestor.ProjectID(projectID),
			bronzegcpbinaryauthorizationattestor.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, att := range staleAttestors {
		if err := s.history.CloseHistory(ctx, tx, att.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for attestor %s: %w", att.ID, err)
		}

		if err := tx.BronzeGCPBinaryAuthorizationAttestor.DeleteOne(att).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete attestor %s: %w", att.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
