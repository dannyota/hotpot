package logbucket

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcploggingbucket"
)

// Service handles GCP Cloud Logging bucket ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new log bucket ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for log bucket ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of log bucket ingestion.
type IngestResult struct {
	ProjectID      string
	BucketCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches log buckets from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	buckets, err := s.client.ListBuckets(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	bucketDataList := make([]*LogBucketData, 0, len(buckets))
	for _, b := range buckets {
		data, err := ConvertLogBucket(b, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert bucket: %w", err)
		}
		bucketDataList = append(bucketDataList, data)
	}

	if err := s.saveBuckets(ctx, bucketDataList); err != nil {
		return nil, fmt.Errorf("failed to save buckets: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		BucketCount:    len(bucketDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveBuckets saves log buckets to the database with history tracking.
func (s *Service) saveBuckets(ctx context.Context, buckets []*LogBucketData) error {
	if len(buckets) == 0 {
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

	for _, bucketData := range buckets {
		existing, err := tx.BronzeGCPLoggingBucket.Query().
			Where(bronzegcploggingbucket.ID(bucketData.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing bucket %s: %w", bucketData.Name, err)
		}

		diff := DiffLogBucketData(existing, bucketData)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPLoggingBucket.UpdateOneID(bucketData.ResourceID).
				SetCollectedAt(bucketData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for bucket %s: %w", bucketData.Name, err)
			}
			continue
		}

		if existing == nil {
			// Create new bucket
			create := tx.BronzeGCPLoggingBucket.Create().
				SetID(bucketData.ResourceID).
				SetName(bucketData.Name).
				SetDescription(bucketData.Description).
				SetRetentionDays(bucketData.RetentionDays).
				SetLocked(bucketData.Locked).
				SetLifecycleState(bucketData.LifecycleState).
				SetAnalyticsEnabled(bucketData.AnalyticsEnabled).
				SetProjectID(bucketData.ProjectID).
				SetLocation(bucketData.Location).
				SetCollectedAt(bucketData.CollectedAt).
				SetFirstCollectedAt(bucketData.CollectedAt)

			if bucketData.CmekSettingsJSON != nil {
				create.SetCmekSettingsJSON(bucketData.CmekSettingsJSON)
			}
			if bucketData.IndexConfigsJSON != nil {
				create.SetIndexConfigsJSON(bucketData.IndexConfigsJSON)
			}

			_, err := create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create bucket %s: %w", bucketData.Name, err)
			}
		} else {
			// Update existing bucket
			update := tx.BronzeGCPLoggingBucket.UpdateOneID(bucketData.ResourceID).
				SetName(bucketData.Name).
				SetDescription(bucketData.Description).
				SetRetentionDays(bucketData.RetentionDays).
				SetLocked(bucketData.Locked).
				SetLifecycleState(bucketData.LifecycleState).
				SetAnalyticsEnabled(bucketData.AnalyticsEnabled).
				SetProjectID(bucketData.ProjectID).
				SetLocation(bucketData.Location).
				SetCollectedAt(bucketData.CollectedAt)

			if bucketData.CmekSettingsJSON != nil {
				update.SetCmekSettingsJSON(bucketData.CmekSettingsJSON)
			}
			if bucketData.IndexConfigsJSON != nil {
				update.SetIndexConfigsJSON(bucketData.IndexConfigsJSON)
			}

			_, err := update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update bucket %s: %w", bucketData.Name, err)
			}
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, bucketData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for bucket %s: %w", bucketData.Name, err)
			}
		} else if diff.IsChanged {
			if err := s.history.UpdateHistory(ctx, tx, existing, bucketData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for bucket %s: %w", bucketData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleBuckets removes log buckets that were not collected in the latest run.
func (s *Service) DeleteStaleBuckets(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGCPLoggingBucket.Query().
		Where(
			bronzegcploggingbucket.ProjectID(projectID),
			bronzegcploggingbucket.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, staleBucket := range stale {
		if err := s.history.CloseHistory(ctx, tx, staleBucket.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for bucket %s: %w", staleBucket.ID, err)
		}
		if err := tx.BronzeGCPLoggingBucket.DeleteOne(staleBucket).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete bucket %s: %w", staleBucket.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
