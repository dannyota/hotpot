package revision

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcprunrevision"
)

// Service handles Cloud Run revision ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new Cloud Run revision ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for Cloud Run revision ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of Cloud Run revision ingestion.
type IngestResult struct {
	ProjectID      string
	RevisionCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches Cloud Run revisions from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch revisions from GCP (queries services from DB, then lists revisions per service)
	rawRevisions, err := s.client.ListRevisions(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list Cloud Run revisions: %w", err)
	}

	// Convert to revision data
	revisionDataList := make([]*RevisionData, 0, len(rawRevisions))
	for _, raw := range rawRevisions {
		data, err := ConvertRevision(raw.ServiceName, raw.Revision, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert Cloud Run revision: %w", err)
		}
		revisionDataList = append(revisionDataList, data)
	}

	// Save to database
	if err := s.saveRevisions(ctx, revisionDataList); err != nil {
		return nil, fmt.Errorf("failed to save Cloud Run revisions: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		RevisionCount:  len(revisionDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveRevisions saves Cloud Run revisions to the database with history tracking.
func (s *Service) saveRevisions(ctx context.Context, revisions []*RevisionData) error {
	if len(revisions) == 0 {
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

	for _, revisionData := range revisions {
		// Load existing revision
		existing, err := tx.BronzeGCPRunRevision.Query().
			Where(bronzegcprunrevision.ID(revisionData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing Cloud Run revision %s: %w", revisionData.ID, err)
		}

		// Compute diff
		diff := DiffRevisionData(existing, revisionData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPRunRevision.UpdateOneID(revisionData.ID).
				SetCollectedAt(revisionData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for Cloud Run revision %s: %w", revisionData.ID, err)
			}
			continue
		}

		// Create or update revision
		if existing == nil {
			create := tx.BronzeGCPRunRevision.Create().
				SetID(revisionData.ID).
				SetName(revisionData.Name).
				SetProjectID(revisionData.ProjectID).
				SetLocation(revisionData.Location).
				SetReconciling(revisionData.Reconciling).
				SetCollectedAt(revisionData.CollectedAt).
				SetFirstCollectedAt(revisionData.CollectedAt).
				SetServiceID(revisionData.ServiceName)

			if revisionData.UID != "" {
				create.SetUID(revisionData.UID)
			}
			if revisionData.Generation != 0 {
				create.SetGeneration(revisionData.Generation)
			}
			if revisionData.LabelsJSON != nil {
				create.SetLabelsJSON(revisionData.LabelsJSON)
			}
			if revisionData.AnnotationsJSON != nil {
				create.SetAnnotationsJSON(revisionData.AnnotationsJSON)
			}
			if revisionData.CreateTime != "" {
				create.SetCreateTime(revisionData.CreateTime)
			}
			if revisionData.UpdateTime != "" {
				create.SetUpdateTime(revisionData.UpdateTime)
			}
			if revisionData.DeleteTime != "" {
				create.SetDeleteTime(revisionData.DeleteTime)
			}
			if revisionData.LaunchStage != 0 {
				create.SetLaunchStage(revisionData.LaunchStage)
			}
			if revisionData.ServiceName != "" {
				create.SetServiceName(revisionData.ServiceName)
			}
			if revisionData.ScalingJSON != nil {
				create.SetScalingJSON(revisionData.ScalingJSON)
			}
			if revisionData.ContainersJSON != nil {
				create.SetContainersJSON(revisionData.ContainersJSON)
			}
			if revisionData.VolumesJSON != nil {
				create.SetVolumesJSON(revisionData.VolumesJSON)
			}
			if revisionData.ExecutionEnvironment != 0 {
				create.SetExecutionEnvironment(revisionData.ExecutionEnvironment)
			}
			if revisionData.EncryptionKey != "" {
				create.SetEncryptionKey(revisionData.EncryptionKey)
			}
			if revisionData.MaxInstanceRequestConcurrency != 0 {
				create.SetMaxInstanceRequestConcurrency(revisionData.MaxInstanceRequestConcurrency)
			}
			if revisionData.Timeout != "" {
				create.SetTimeout(revisionData.Timeout)
			}
			if revisionData.ServiceAccount != "" {
				create.SetServiceAccount(revisionData.ServiceAccount)
			}
			if revisionData.ConditionsJSON != nil {
				create.SetConditionsJSON(revisionData.ConditionsJSON)
			}
			if revisionData.ObservedGeneration != 0 {
				create.SetObservedGeneration(revisionData.ObservedGeneration)
			}
			if revisionData.LogURI != "" {
				create.SetLogURI(revisionData.LogURI)
			}
			if revisionData.Etag != "" {
				create.SetEtag(revisionData.Etag)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create Cloud Run revision %s: %w", revisionData.ID, err)
			}
		} else {
			update := tx.BronzeGCPRunRevision.UpdateOneID(revisionData.ID).
				SetName(revisionData.Name).
				SetProjectID(revisionData.ProjectID).
				SetLocation(revisionData.Location).
				SetReconciling(revisionData.Reconciling).
				SetCollectedAt(revisionData.CollectedAt)

			if revisionData.UID != "" {
				update.SetUID(revisionData.UID)
			}
			if revisionData.Generation != 0 {
				update.SetGeneration(revisionData.Generation)
			}
			if revisionData.LabelsJSON != nil {
				update.SetLabelsJSON(revisionData.LabelsJSON)
			}
			if revisionData.AnnotationsJSON != nil {
				update.SetAnnotationsJSON(revisionData.AnnotationsJSON)
			}
			if revisionData.CreateTime != "" {
				update.SetCreateTime(revisionData.CreateTime)
			}
			if revisionData.UpdateTime != "" {
				update.SetUpdateTime(revisionData.UpdateTime)
			}
			if revisionData.DeleteTime != "" {
				update.SetDeleteTime(revisionData.DeleteTime)
			}
			if revisionData.LaunchStage != 0 {
				update.SetLaunchStage(revisionData.LaunchStage)
			}
			if revisionData.ServiceName != "" {
				update.SetServiceName(revisionData.ServiceName)
			}
			if revisionData.ScalingJSON != nil {
				update.SetScalingJSON(revisionData.ScalingJSON)
			}
			if revisionData.ContainersJSON != nil {
				update.SetContainersJSON(revisionData.ContainersJSON)
			}
			if revisionData.VolumesJSON != nil {
				update.SetVolumesJSON(revisionData.VolumesJSON)
			}
			if revisionData.ExecutionEnvironment != 0 {
				update.SetExecutionEnvironment(revisionData.ExecutionEnvironment)
			}
			if revisionData.EncryptionKey != "" {
				update.SetEncryptionKey(revisionData.EncryptionKey)
			}
			if revisionData.MaxInstanceRequestConcurrency != 0 {
				update.SetMaxInstanceRequestConcurrency(revisionData.MaxInstanceRequestConcurrency)
			}
			if revisionData.Timeout != "" {
				update.SetTimeout(revisionData.Timeout)
			}
			if revisionData.ServiceAccount != "" {
				update.SetServiceAccount(revisionData.ServiceAccount)
			}
			if revisionData.ConditionsJSON != nil {
				update.SetConditionsJSON(revisionData.ConditionsJSON)
			}
			if revisionData.ObservedGeneration != 0 {
				update.SetObservedGeneration(revisionData.ObservedGeneration)
			}
			if revisionData.LogURI != "" {
				update.SetLogURI(revisionData.LogURI)
			}
			if revisionData.Etag != "" {
				update.SetEtag(revisionData.Etag)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update Cloud Run revision %s: %w", revisionData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, revisionData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for Cloud Run revision %s: %w", revisionData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, revisionData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for Cloud Run revision %s: %w", revisionData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleRevisions removes revisions that were not collected in the latest run.
func (s *Service) DeleteStaleRevisions(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	staleRevisions, err := tx.BronzeGCPRunRevision.Query().
		Where(
			bronzegcprunrevision.ProjectID(projectID),
			bronzegcprunrevision.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, rev := range staleRevisions {
		if err := s.history.CloseHistory(ctx, tx, rev.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for Cloud Run revision %s: %w", rev.ID, err)
		}

		if err := tx.BronzeGCPRunRevision.DeleteOne(rev).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete Cloud Run revision %s: %w", rev.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
