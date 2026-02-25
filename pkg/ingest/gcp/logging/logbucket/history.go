package logbucket

import (
	"context"
	"fmt"
	"time"

	entlogging "github.com/dannyota/hotpot/pkg/storage/ent/gcp/logging"
	"github.com/dannyota/hotpot/pkg/storage/ent/gcp/logging/bronzehistorygcploggingbucket"
)

// HistoryService handles history tracking for log buckets.
type HistoryService struct {
	entClient *entlogging.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entlogging.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new log bucket.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entlogging.Tx, bucketData *LogBucketData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPLoggingBucket.Create().
		SetResourceID(bucketData.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(bucketData.CollectedAt).
		SetFirstCollectedAt(bucketData.CollectedAt).
		SetName(bucketData.Name).
		SetDescription(bucketData.Description).
		SetRetentionDays(bucketData.RetentionDays).
		SetLocked(bucketData.Locked).
		SetLifecycleState(bucketData.LifecycleState).
		SetAnalyticsEnabled(bucketData.AnalyticsEnabled).
		SetProjectID(bucketData.ProjectID).
		SetLocation(bucketData.Location).
		SetCmekSettingsJSON(bucketData.CmekSettingsJSON).
		SetIndexConfigsJSON(bucketData.IndexConfigsJSON).
		Save(ctx)
	return err
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entlogging.Tx, old *entlogging.BronzeGCPLoggingBucket, new *LogBucketData, diff *LogBucketDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	_, err := tx.BronzeHistoryGCPLoggingBucket.Update().
		Where(
			bronzehistorygcploggingbucket.ResourceID(old.ID),
			bronzehistorygcploggingbucket.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close old history: %w", err)
	}

	// Create new history
	_, err = tx.BronzeHistoryGCPLoggingBucket.Create().
		SetResourceID(new.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetDescription(new.Description).
		SetRetentionDays(new.RetentionDays).
		SetLocked(new.Locked).
		SetLifecycleState(new.LifecycleState).
		SetAnalyticsEnabled(new.AnalyticsEnabled).
		SetProjectID(new.ProjectID).
		SetLocation(new.Location).
		SetCmekSettingsJSON(new.CmekSettingsJSON).
		SetIndexConfigsJSON(new.IndexConfigsJSON).
		Save(ctx)
	return err
}

// CloseHistory closes history records for a deleted log bucket.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entlogging.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPLoggingBucket.Update().
		Where(
			bronzehistorygcploggingbucket.ResourceID(resourceID),
			bronzehistorygcploggingbucket.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if entlogging.IsNotFound(err) {
		return nil // No history to close
	}
	return err
}
