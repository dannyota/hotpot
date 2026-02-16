package instance

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpspannerinstance"
)

// HistoryService manages Spanner instance history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new Spanner instance.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *InstanceData, now time.Time) error {
	create := tx.BronzeHistoryGCPSpannerInstance.Create().
		SetResourceID(data.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetConfig(data.Config).
		SetDisplayName(data.DisplayName).
		SetNodeCount(data.NodeCount).
		SetProcessingUnits(data.ProcessingUnits).
		SetState(data.State).
		SetCreateTime(data.CreateTime).
		SetUpdateTime(data.UpdateTime).
		SetEdition(data.Edition).
		SetDefaultBackupScheduleType(data.DefaultBackupScheduleType).
		SetProjectID(data.ProjectID)

	if data.LabelsJSON != nil {
		create.SetLabelsJSON(data.LabelsJSON)
	}
	if data.EndpointUrisJSON != nil {
		create.SetEndpointUrisJSON(data.EndpointUrisJSON)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Spanner instance history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed Spanner instance.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPSpannerInstance, new *InstanceData, diff *InstanceDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPSpannerInstance.Query().
		Where(
			bronzehistorygcpspannerinstance.ResourceID(old.ID),
			bronzehistorygcpspannerinstance.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current Spanner instance history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPSpannerInstance.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current Spanner instance history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPSpannerInstance.Create().
			SetResourceID(new.ResourceID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetConfig(new.Config).
			SetDisplayName(new.DisplayName).
			SetNodeCount(new.NodeCount).
			SetProcessingUnits(new.ProcessingUnits).
			SetState(new.State).
			SetCreateTime(new.CreateTime).
			SetUpdateTime(new.UpdateTime).
			SetEdition(new.Edition).
			SetDefaultBackupScheduleType(new.DefaultBackupScheduleType).
			SetProjectID(new.ProjectID)

		if new.LabelsJSON != nil {
			create.SetLabelsJSON(new.LabelsJSON)
		}
		if new.EndpointUrisJSON != nil {
			create.SetEndpointUrisJSON(new.EndpointUrisJSON)
		}

		_, err = create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new Spanner instance history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted Spanner instance.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPSpannerInstance.Query().
		Where(
			bronzehistorygcpspannerinstance.ResourceID(resourceID),
			bronzehistorygcpspannerinstance.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current Spanner instance history: %w", err)
	}

	err = tx.BronzeHistoryGCPSpannerInstance.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close Spanner instance history: %w", err)
	}

	return nil
}
