package instance

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpbigtableinstance"
)

// HistoryService manages Bigtable instance history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new Bigtable instance.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *InstanceData, now time.Time) error {
	create := tx.BronzeHistoryGCPBigtableInstance.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetDisplayName(data.DisplayName).
		SetState(data.State).
		SetInstanceType(data.InstanceType).
		SetProjectID(data.ProjectID)

	if data.CreateTime != "" {
		create.SetCreateTime(data.CreateTime)
	}
	if data.SatisfiesPzs != nil {
		create.SetSatisfiesPzs(*data.SatisfiesPzs)
	}
	if data.LabelsJSON != nil {
		create.SetLabelsJSON(data.LabelsJSON)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Bigtable instance history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed Bigtable instance.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPBigtableInstance, new *InstanceData, diff *InstanceDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPBigtableInstance.Query().
		Where(
			bronzehistorygcpbigtableinstance.ResourceID(old.ID),
			bronzehistorygcpbigtableinstance.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current Bigtable instance history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPBigtableInstance.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current Bigtable instance history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPBigtableInstance.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetDisplayName(new.DisplayName).
			SetState(new.State).
			SetInstanceType(new.InstanceType).
			SetProjectID(new.ProjectID)

		if new.CreateTime != "" {
			create.SetCreateTime(new.CreateTime)
		}
		if new.SatisfiesPzs != nil {
			create.SetSatisfiesPzs(*new.SatisfiesPzs)
		}
		if new.LabelsJSON != nil {
			create.SetLabelsJSON(new.LabelsJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new Bigtable instance history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted Bigtable instance.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPBigtableInstance.Query().
		Where(
			bronzehistorygcpbigtableinstance.ResourceID(resourceID),
			bronzehistorygcpbigtableinstance.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current Bigtable instance history: %w", err)
	}

	err = tx.BronzeHistoryGCPBigtableInstance.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close Bigtable instance history: %w", err)
	}

	return nil
}
