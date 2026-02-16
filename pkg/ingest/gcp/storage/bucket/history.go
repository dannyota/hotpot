package bucket

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpstoragebucket"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpstoragebucketlabel"
)

// HistoryService handles history tracking for buckets.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new bucket and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, bucketData *BucketData, now time.Time) error {
	hist, err := tx.BronzeHistoryGCPStorageBucket.Create().
		SetResourceID(bucketData.ID).
		SetValidFrom(now).
		SetCollectedAt(bucketData.CollectedAt).
		SetFirstCollectedAt(bucketData.CollectedAt).
		SetName(bucketData.Name).
		SetLocation(bucketData.Location).
		SetStorageClass(bucketData.StorageClass).
		SetProjectNumber(bucketData.ProjectNumber).
		SetTimeCreated(bucketData.TimeCreated).
		SetUpdated(bucketData.Updated).
		SetDefaultEventBasedHold(bucketData.DefaultEventBasedHold).
		SetMetageneration(bucketData.Metageneration).
		SetEtag(bucketData.Etag).
		SetIamConfigurationJSON(bucketData.IamConfigurationJSON).
		SetEncryptionJSON(bucketData.EncryptionJSON).
		SetLifecycleJSON(bucketData.LifecycleJSON).
		SetVersioningJSON(bucketData.VersioningJSON).
		SetRetentionPolicyJSON(bucketData.RetentionPolicyJSON).
		SetLoggingJSON(bucketData.LoggingJSON).
		SetCorsJSON(bucketData.CorsJSON).
		SetWebsiteJSON(bucketData.WebsiteJSON).
		SetAutoclassJSON(bucketData.AutoclassJSON).
		SetProjectID(bucketData.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create bucket history: %w", err)
	}

	return h.createLabelsHistory(ctx, tx, hist.HistoryID, bucketData, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPStorageBucket, new *BucketData, diff *BucketDiff, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGCPStorageBucket.Query().
		Where(
			bronzehistorygcpstoragebucket.ResourceID(old.ID),
			bronzehistorygcpstoragebucket.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current bucket history: %w", err)
	}

	if diff.IsChanged {
		err = tx.BronzeHistoryGCPStorageBucket.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current bucket history: %w", err)
		}

		hist, err := tx.BronzeHistoryGCPStorageBucket.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetLocation(new.Location).
			SetStorageClass(new.StorageClass).
			SetProjectNumber(new.ProjectNumber).
			SetTimeCreated(new.TimeCreated).
			SetUpdated(new.Updated).
			SetDefaultEventBasedHold(new.DefaultEventBasedHold).
			SetMetageneration(new.Metageneration).
			SetEtag(new.Etag).
			SetIamConfigurationJSON(new.IamConfigurationJSON).
			SetEncryptionJSON(new.EncryptionJSON).
			SetLifecycleJSON(new.LifecycleJSON).
			SetVersioningJSON(new.VersioningJSON).
			SetRetentionPolicyJSON(new.RetentionPolicyJSON).
			SetLoggingJSON(new.LoggingJSON).
			SetCorsJSON(new.CorsJSON).
			SetWebsiteJSON(new.WebsiteJSON).
			SetAutoclassJSON(new.AutoclassJSON).
			SetProjectID(new.ProjectID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new bucket history: %w", err)
		}

		if err := h.closeLabelsHistory(ctx, tx, currentHist.HistoryID, now); err != nil {
			return fmt.Errorf("failed to close labels history: %w", err)
		}
		return h.createLabelsHistory(ctx, tx, hist.HistoryID, new, now)
	}

	if diff.LabelDiff.Changed {
		return h.updateLabelsHistory(ctx, tx, currentHist.HistoryID, new, now)
	}

	return nil
}

// CloseHistory closes history records for a deleted bucket.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGCPStorageBucket.Query().
		Where(
			bronzehistorygcpstoragebucket.ResourceID(resourceID),
			bronzehistorygcpstoragebucket.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current bucket history: %w", err)
	}

	err = tx.BronzeHistoryGCPStorageBucket.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close bucket history: %w", err)
	}

	return h.closeLabelsHistory(ctx, tx, currentHist.HistoryID, now)
}

func (h *HistoryService) createLabelsHistory(ctx context.Context, tx *ent.Tx, bucketHistoryID uint, bucketData *BucketData, now time.Time) error {
	for _, label := range bucketData.Labels {
		_, err := tx.BronzeHistoryGCPStorageBucketLabel.Create().
			SetBucketHistoryID(bucketHistoryID).
			SetValidFrom(now).
			SetKey(label.Key).
			SetValue(label.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label history: %w", err)
		}
	}
	return nil
}

func (h *HistoryService) closeLabelsHistory(ctx context.Context, tx *ent.Tx, bucketHistoryID uint, now time.Time) error {
	_, err := tx.BronzeHistoryGCPStorageBucketLabel.Update().
		Where(
			bronzehistorygcpstoragebucketlabel.BucketHistoryID(bucketHistoryID),
			bronzehistorygcpstoragebucketlabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close label history: %w", err)
	}
	return nil
}

func (h *HistoryService) updateLabelsHistory(ctx context.Context, tx *ent.Tx, bucketHistoryID uint, new *BucketData, now time.Time) error {
	if err := h.closeLabelsHistory(ctx, tx, bucketHistoryID, now); err != nil {
		return err
	}
	return h.createLabelsHistory(ctx, tx, bucketHistoryID, new, now)
}
