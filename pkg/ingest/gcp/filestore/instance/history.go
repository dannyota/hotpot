package instance

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpfilestoreinstance"
)

// HistoryService manages Filestore instance history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new Filestore instance.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *InstanceData, now time.Time) error {
	create := tx.BronzeHistoryGCPFilestoreInstance.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetDescription(data.Description).
		SetState(data.State).
		SetStatusMessage(data.StatusMessage).
		SetCreateTime(data.CreateTime).
		SetTier(data.Tier).
		SetEtag(data.Etag).
		SetSatisfiesPzs(data.SatisfiesPzs).
		SetSatisfiesPzi(data.SatisfiesPzi).
		SetKmsKeyName(data.KmsKeyName).
		SetMaxCapacityGB(data.MaxCapacityGB).
		SetProtocol(data.Protocol).
		SetProjectID(data.ProjectID).
		SetLocation(data.Location)

	if data.LabelsJSON != nil {
		create.SetLabelsJSON(data.LabelsJSON)
	}
	if data.FileSharesJSON != nil {
		create.SetFileSharesJSON(data.FileSharesJSON)
	}
	if data.NetworksJSON != nil {
		create.SetNetworksJSON(data.NetworksJSON)
	}
	if data.SuspensionReasonsJSON != nil {
		create.SetSuspensionReasonsJSON(data.SuspensionReasonsJSON)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Filestore instance history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed Filestore instance.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPFilestoreInstance, new *InstanceData, diff *InstanceDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPFilestoreInstance.Query().
		Where(
			bronzehistorygcpfilestoreinstance.ResourceID(old.ID),
			bronzehistorygcpfilestoreinstance.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current Filestore instance history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPFilestoreInstance.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current Filestore instance history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPFilestoreInstance.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetDescription(new.Description).
			SetState(new.State).
			SetStatusMessage(new.StatusMessage).
			SetCreateTime(new.CreateTime).
			SetTier(new.Tier).
			SetEtag(new.Etag).
			SetSatisfiesPzs(new.SatisfiesPzs).
			SetSatisfiesPzi(new.SatisfiesPzi).
			SetKmsKeyName(new.KmsKeyName).
			SetMaxCapacityGB(new.MaxCapacityGB).
			SetProtocol(new.Protocol).
			SetProjectID(new.ProjectID).
			SetLocation(new.Location)

		if new.LabelsJSON != nil {
			create.SetLabelsJSON(new.LabelsJSON)
		}
		if new.FileSharesJSON != nil {
			create.SetFileSharesJSON(new.FileSharesJSON)
		}
		if new.NetworksJSON != nil {
			create.SetNetworksJSON(new.NetworksJSON)
		}
		if new.SuspensionReasonsJSON != nil {
			create.SetSuspensionReasonsJSON(new.SuspensionReasonsJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new Filestore instance history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted Filestore instance.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPFilestoreInstance.Query().
		Where(
			bronzehistorygcpfilestoreinstance.ResourceID(resourceID),
			bronzehistorygcpfilestoreinstance.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current Filestore instance history: %w", err)
	}

	err = tx.BronzeHistoryGCPFilestoreInstance.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close Filestore instance history: %w", err)
	}

	return nil
}
