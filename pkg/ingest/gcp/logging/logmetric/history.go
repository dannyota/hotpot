package logmetric

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcplogginglogmetric"
)

// HistoryService handles history tracking for log metrics.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new log metric.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *LogMetricData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPLoggingLogMetric.Create().
		SetResourceID(data.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetDescription(data.Description).
		SetFilter(data.Filter).
		SetMetricDescriptorJSON(data.MetricDescriptorJSON).
		SetLabelExtractorsJSON(data.LabelExtractorsJSON).
		SetBucketOptionsJSON(data.BucketOptionsJSON).
		SetValueExtractor(data.ValueExtractor).
		SetVersion(data.Version).
		SetDisabled(data.Disabled).
		SetCreateTime(data.CreateTime).
		SetUpdateTime(data.UpdateTime).
		SetProjectID(data.ProjectID).
		Save(ctx)
	return err
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPLoggingLogMetric, new *LogMetricData, diff *LogMetricDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	_, err := tx.BronzeHistoryGCPLoggingLogMetric.Update().
		Where(
			bronzehistorygcplogginglogmetric.ResourceID(old.ID),
			bronzehistorygcplogginglogmetric.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close old history: %w", err)
	}

	// Create new history
	_, err = tx.BronzeHistoryGCPLoggingLogMetric.Create().
		SetResourceID(new.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetDescription(new.Description).
		SetFilter(new.Filter).
		SetMetricDescriptorJSON(new.MetricDescriptorJSON).
		SetLabelExtractorsJSON(new.LabelExtractorsJSON).
		SetBucketOptionsJSON(new.BucketOptionsJSON).
		SetValueExtractor(new.ValueExtractor).
		SetVersion(new.Version).
		SetDisabled(new.Disabled).
		SetCreateTime(new.CreateTime).
		SetUpdateTime(new.UpdateTime).
		SetProjectID(new.ProjectID).
		Save(ctx)
	return err
}

// CloseHistory closes history records for a deleted log metric.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPLoggingLogMetric.Update().
		Where(
			bronzehistorygcplogginglogmetric.ResourceID(resourceID),
			bronzehistorygcplogginglogmetric.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if ent.IsNotFound(err) {
		return nil // No history to close
	}
	return err
}
