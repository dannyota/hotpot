package logexclusion

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcplogginglogexclusion"
)

// HistoryService handles history tracking for log exclusions.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new log exclusion.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *LogExclusionData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPLoggingLogExclusion.Create().
		SetResourceID(data.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetDescription(data.Description).
		SetFilter(data.Filter).
		SetDisabled(data.Disabled).
		SetCreateTime(data.CreateTime).
		SetUpdateTime(data.UpdateTime).
		SetProjectID(data.ProjectID).
		Save(ctx)
	return err
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPLoggingLogExclusion, new *LogExclusionData, diff *ExclusionDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	_, err := tx.BronzeHistoryGCPLoggingLogExclusion.Update().
		Where(
			bronzehistorygcplogginglogexclusion.ResourceID(old.ID),
			bronzehistorygcplogginglogexclusion.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close old history: %w", err)
	}

	// Create new history
	_, err = tx.BronzeHistoryGCPLoggingLogExclusion.Create().
		SetResourceID(new.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetDescription(new.Description).
		SetFilter(new.Filter).
		SetDisabled(new.Disabled).
		SetCreateTime(new.CreateTime).
		SetUpdateTime(new.UpdateTime).
		SetProjectID(new.ProjectID).
		Save(ctx)
	return err
}

// CloseHistory closes history records for a deleted log exclusion.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPLoggingLogExclusion.Update().
		Where(
			bronzehistorygcplogginglogexclusion.ResourceID(resourceID),
			bronzehistorygcplogginglogexclusion.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if ent.IsNotFound(err) {
		return nil // No history to close
	}
	return err
}
