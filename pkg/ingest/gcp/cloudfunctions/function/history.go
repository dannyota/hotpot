package function

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcloudfunctionsfunction"
)

// HistoryService manages Cloud Function history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new Cloud Function.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *FunctionData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPCloudFunctionsFunction.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetDescription(data.Description).
		SetEnvironment(data.Environment).
		SetState(data.State).
		SetBuildConfigJSON(data.BuildConfigJSON).
		SetServiceConfigJSON(data.ServiceConfigJSON).
		SetEventTriggerJSON(data.EventTriggerJSON).
		SetStateMessagesJSON(data.StateMessagesJSON).
		SetUpdateTime(data.UpdateTime).
		SetCreateTime(data.CreateTime).
		SetLabelsJSON(data.LabelsJSON).
		SetKmsKeyName(data.KmsKeyName).
		SetURL(data.URL).
		SetSatisfiesPzs(data.SatisfiesPzs).
		SetProjectID(data.ProjectID).
		SetLocation(data.Location).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Cloud Function history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed Cloud Function.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPCloudFunctionsFunction, new *FunctionData, diff *FunctionDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPCloudFunctionsFunction.Query().
		Where(
			bronzehistorygcpcloudfunctionsfunction.ResourceID(old.ID),
			bronzehistorygcpcloudfunctionsfunction.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current Cloud Function history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPCloudFunctionsFunction.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current Cloud Function history: %w", err)
		}

		// Create new history
		_, err := tx.BronzeHistoryGCPCloudFunctionsFunction.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetDescription(new.Description).
			SetEnvironment(new.Environment).
			SetState(new.State).
			SetBuildConfigJSON(new.BuildConfigJSON).
			SetServiceConfigJSON(new.ServiceConfigJSON).
			SetEventTriggerJSON(new.EventTriggerJSON).
			SetStateMessagesJSON(new.StateMessagesJSON).
			SetUpdateTime(new.UpdateTime).
			SetCreateTime(new.CreateTime).
			SetLabelsJSON(new.LabelsJSON).
			SetKmsKeyName(new.KmsKeyName).
			SetURL(new.URL).
			SetSatisfiesPzs(new.SatisfiesPzs).
			SetProjectID(new.ProjectID).
			SetLocation(new.Location).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new Cloud Function history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted Cloud Function.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPCloudFunctionsFunction.Query().
		Where(
			bronzehistorygcpcloudfunctionsfunction.ResourceID(resourceID),
			bronzehistorygcpcloudfunctionsfunction.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current Cloud Function history: %w", err)
	}

	err = tx.BronzeHistoryGCPCloudFunctionsFunction.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close Cloud Function history: %w", err)
	}

	return nil
}
