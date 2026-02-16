package sink

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcploggingsink"
)

// HistoryService handles history tracking for sinks.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new sink.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, sinkData *SinkData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPLoggingSink.Create().
		SetResourceID(sinkData.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(sinkData.CollectedAt).
		SetFirstCollectedAt(sinkData.CollectedAt).
		SetName(sinkData.Name).
		SetDestination(sinkData.Destination).
		SetFilter(sinkData.Filter).
		SetDescription(sinkData.Description).
		SetDisabled(sinkData.Disabled).
		SetIncludeChildren(sinkData.IncludeChildren).
		SetWriterIdentity(sinkData.WriterIdentity).
		SetExclusionsJSON(sinkData.ExclusionsJSON).
		SetBigqueryOptionsJSON(sinkData.BigqueryOptionsJSON).
		SetProjectID(sinkData.ProjectID).
		Save(ctx)
	return err
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPLoggingSink, new *SinkData, diff *SinkDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	_, err := tx.BronzeHistoryGCPLoggingSink.Update().
		Where(
			bronzehistorygcploggingsink.ResourceID(old.ID),
			bronzehistorygcploggingsink.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close old history: %w", err)
	}

	// Create new history
	_, err = tx.BronzeHistoryGCPLoggingSink.Create().
		SetResourceID(new.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetDestination(new.Destination).
		SetFilter(new.Filter).
		SetDescription(new.Description).
		SetDisabled(new.Disabled).
		SetIncludeChildren(new.IncludeChildren).
		SetWriterIdentity(new.WriterIdentity).
		SetExclusionsJSON(new.ExclusionsJSON).
		SetBigqueryOptionsJSON(new.BigqueryOptionsJSON).
		SetProjectID(new.ProjectID).
		Save(ctx)
	return err
}

// CloseHistory closes history records for a deleted sink.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPLoggingSink.Update().
		Where(
			bronzehistorygcploggingsink.ResourceID(resourceID),
			bronzehistorygcploggingsink.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if ent.IsNotFound(err) {
		return nil // No history to close
	}
	return err
}
