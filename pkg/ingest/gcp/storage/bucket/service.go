package bucket

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpstoragebucket"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpstoragebucketlabel"
)

// Service handles GCP Storage bucket ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new bucket ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for bucket ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of bucket ingestion.
type IngestResult struct {
	ProjectID      string
	BucketCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches buckets from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	buckets, err := s.client.ListBuckets(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	bucketDataList := make([]*BucketData, 0, len(buckets))
	for _, b := range buckets {
		data, err := ConvertBucket(b, params.ProjectID, collectedAt)
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

func (s *Service) saveBuckets(ctx context.Context, buckets []*BucketData) error {
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
		existing, err := tx.BronzeGCPStorageBucket.Query().
			Where(bronzegcpstoragebucket.ID(bucketData.ID)).
			WithLabels().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing bucket %s: %w", bucketData.Name, err)
		}

		diff := DiffBucketData(existing, bucketData)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPStorageBucket.UpdateOneID(bucketData.ID).
				SetCollectedAt(bucketData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for bucket %s: %w", bucketData.Name, err)
			}
			continue
		}

		// Delete old children if updating
		if existing != nil {
			if err := deleteBucketChildren(ctx, tx, bucketData.ID); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old children for bucket %s: %w", bucketData.Name, err)
			}
		}

		var savedBucket *ent.BronzeGCPStorageBucket
		if existing == nil {
			create := tx.BronzeGCPStorageBucket.Create().
				SetID(bucketData.ID).
				SetName(bucketData.Name).
				SetLocation(bucketData.Location).
				SetStorageClass(bucketData.StorageClass).
				SetProjectNumber(bucketData.ProjectNumber).
				SetTimeCreated(bucketData.TimeCreated).
				SetUpdated(bucketData.Updated).
				SetDefaultEventBasedHold(bucketData.DefaultEventBasedHold).
				SetMetageneration(bucketData.Metageneration).
				SetEtag(bucketData.Etag).
				SetProjectID(bucketData.ProjectID).
				SetCollectedAt(bucketData.CollectedAt).
				SetFirstCollectedAt(bucketData.CollectedAt)

			if bucketData.IamConfigurationJSON != nil {
				create.SetIamConfigurationJSON(bucketData.IamConfigurationJSON)
			}
			if bucketData.EncryptionJSON != nil {
				create.SetEncryptionJSON(bucketData.EncryptionJSON)
			}
			if bucketData.LifecycleJSON != nil {
				create.SetLifecycleJSON(bucketData.LifecycleJSON)
			}
			if bucketData.VersioningJSON != nil {
				create.SetVersioningJSON(bucketData.VersioningJSON)
			}
			if bucketData.RetentionPolicyJSON != nil {
				create.SetRetentionPolicyJSON(bucketData.RetentionPolicyJSON)
			}
			if bucketData.LoggingJSON != nil {
				create.SetLoggingJSON(bucketData.LoggingJSON)
			}
			if bucketData.CorsJSON != nil {
				create.SetCorsJSON(bucketData.CorsJSON)
			}
			if bucketData.WebsiteJSON != nil {
				create.SetWebsiteJSON(bucketData.WebsiteJSON)
			}
			if bucketData.AutoclassJSON != nil {
				create.SetAutoclassJSON(bucketData.AutoclassJSON)
			}

			savedBucket, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create bucket %s: %w", bucketData.Name, err)
			}
		} else {
			update := tx.BronzeGCPStorageBucket.UpdateOneID(bucketData.ID).
				SetName(bucketData.Name).
				SetLocation(bucketData.Location).
				SetStorageClass(bucketData.StorageClass).
				SetProjectNumber(bucketData.ProjectNumber).
				SetTimeCreated(bucketData.TimeCreated).
				SetUpdated(bucketData.Updated).
				SetDefaultEventBasedHold(bucketData.DefaultEventBasedHold).
				SetMetageneration(bucketData.Metageneration).
				SetEtag(bucketData.Etag).
				SetProjectID(bucketData.ProjectID).
				SetCollectedAt(bucketData.CollectedAt)

			if bucketData.IamConfigurationJSON != nil {
				update.SetIamConfigurationJSON(bucketData.IamConfigurationJSON)
			}
			if bucketData.EncryptionJSON != nil {
				update.SetEncryptionJSON(bucketData.EncryptionJSON)
			}
			if bucketData.LifecycleJSON != nil {
				update.SetLifecycleJSON(bucketData.LifecycleJSON)
			}
			if bucketData.VersioningJSON != nil {
				update.SetVersioningJSON(bucketData.VersioningJSON)
			}
			if bucketData.RetentionPolicyJSON != nil {
				update.SetRetentionPolicyJSON(bucketData.RetentionPolicyJSON)
			}
			if bucketData.LoggingJSON != nil {
				update.SetLoggingJSON(bucketData.LoggingJSON)
			}
			if bucketData.CorsJSON != nil {
				update.SetCorsJSON(bucketData.CorsJSON)
			}
			if bucketData.WebsiteJSON != nil {
				update.SetWebsiteJSON(bucketData.WebsiteJSON)
			}
			if bucketData.AutoclassJSON != nil {
				update.SetAutoclassJSON(bucketData.AutoclassJSON)
			}

			savedBucket, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update bucket %s: %w", bucketData.Name, err)
			}
		}

		if err := createBucketChildren(ctx, tx, savedBucket, bucketData); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create children for bucket %s: %w", bucketData.Name, err)
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, bucketData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for bucket %s: %w", bucketData.Name, err)
			}
		} else {
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

func deleteBucketChildren(ctx context.Context, tx *ent.Tx, bucketID string) error {
	_, err := tx.BronzeGCPStorageBucketLabel.Delete().
		Where(bronzegcpstoragebucketlabel.HasBucketWith(bronzegcpstoragebucket.ID(bucketID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete labels: %w", err)
	}
	return nil
}

func createBucketChildren(ctx context.Context, tx *ent.Tx, savedBucket *ent.BronzeGCPStorageBucket, bucketData *BucketData) error {
	for _, label := range bucketData.Labels {
		_, err := tx.BronzeGCPStorageBucketLabel.Create().
			SetKey(label.Key).
			SetValue(label.Value).
			SetBucket(savedBucket).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label: %w", err)
		}
	}
	return nil
}

// DeleteStaleBuckets removes buckets that were not collected in the latest run.
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

	staleBuckets, err := tx.BronzeGCPStorageBucket.Query().
		Where(
			bronzegcpstoragebucket.ProjectID(projectID),
			bronzegcpstoragebucket.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, b := range staleBuckets {
		if err := s.history.CloseHistory(ctx, tx, b.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for bucket %s: %w", b.ID, err)
		}

		if err := tx.BronzeGCPStorageBucket.DeleteOne(b).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete bucket %s: %w", b.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
