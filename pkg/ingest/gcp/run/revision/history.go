package revision

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcprunrevision"
)

// HistoryService manages Cloud Run revision history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new Cloud Run revision.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *RevisionData, now time.Time) error {
	create := tx.BronzeHistoryGCPRunRevision.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetProjectID(data.ProjectID).
		SetLocation(data.Location).
		SetReconciling(data.Reconciling)

	if data.UID != "" {
		create.SetUID(data.UID)
	}
	if data.Generation != 0 {
		create.SetGeneration(data.Generation)
	}
	if data.LabelsJSON != nil {
		create.SetLabelsJSON(data.LabelsJSON)
	}
	if data.AnnotationsJSON != nil {
		create.SetAnnotationsJSON(data.AnnotationsJSON)
	}
	if data.CreateTime != "" {
		create.SetCreateTime(data.CreateTime)
	}
	if data.UpdateTime != "" {
		create.SetUpdateTime(data.UpdateTime)
	}
	if data.DeleteTime != "" {
		create.SetDeleteTime(data.DeleteTime)
	}
	if data.LaunchStage != 0 {
		create.SetLaunchStage(data.LaunchStage)
	}
	if data.ServiceName != "" {
		create.SetServiceName(data.ServiceName)
	}
	if data.ScalingJSON != nil {
		create.SetScalingJSON(data.ScalingJSON)
	}
	if data.ContainersJSON != nil {
		create.SetContainersJSON(data.ContainersJSON)
	}
	if data.VolumesJSON != nil {
		create.SetVolumesJSON(data.VolumesJSON)
	}
	if data.ExecutionEnvironment != 0 {
		create.SetExecutionEnvironment(data.ExecutionEnvironment)
	}
	if data.EncryptionKey != "" {
		create.SetEncryptionKey(data.EncryptionKey)
	}
	if data.MaxInstanceRequestConcurrency != 0 {
		create.SetMaxInstanceRequestConcurrency(data.MaxInstanceRequestConcurrency)
	}
	if data.Timeout != "" {
		create.SetTimeout(data.Timeout)
	}
	if data.ServiceAccount != "" {
		create.SetServiceAccount(data.ServiceAccount)
	}
	if data.ConditionsJSON != nil {
		create.SetConditionsJSON(data.ConditionsJSON)
	}
	if data.ObservedGeneration != 0 {
		create.SetObservedGeneration(data.ObservedGeneration)
	}
	if data.LogURI != "" {
		create.SetLogURI(data.LogURI)
	}
	if data.Etag != "" {
		create.SetEtag(data.Etag)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Cloud Run revision history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed Cloud Run revision.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPRunRevision, new *RevisionData, diff *RevisionDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPRunRevision.Query().
		Where(
			bronzehistorygcprunrevision.ResourceID(old.ID),
			bronzehistorygcprunrevision.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current Cloud Run revision history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPRunRevision.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current Cloud Run revision history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPRunRevision.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetProjectID(new.ProjectID).
			SetLocation(new.Location).
			SetReconciling(new.Reconciling)

		if new.UID != "" {
			create.SetUID(new.UID)
		}
		if new.Generation != 0 {
			create.SetGeneration(new.Generation)
		}
		if new.LabelsJSON != nil {
			create.SetLabelsJSON(new.LabelsJSON)
		}
		if new.AnnotationsJSON != nil {
			create.SetAnnotationsJSON(new.AnnotationsJSON)
		}
		if new.CreateTime != "" {
			create.SetCreateTime(new.CreateTime)
		}
		if new.UpdateTime != "" {
			create.SetUpdateTime(new.UpdateTime)
		}
		if new.DeleteTime != "" {
			create.SetDeleteTime(new.DeleteTime)
		}
		if new.LaunchStage != 0 {
			create.SetLaunchStage(new.LaunchStage)
		}
		if new.ServiceName != "" {
			create.SetServiceName(new.ServiceName)
		}
		if new.ScalingJSON != nil {
			create.SetScalingJSON(new.ScalingJSON)
		}
		if new.ContainersJSON != nil {
			create.SetContainersJSON(new.ContainersJSON)
		}
		if new.VolumesJSON != nil {
			create.SetVolumesJSON(new.VolumesJSON)
		}
		if new.ExecutionEnvironment != 0 {
			create.SetExecutionEnvironment(new.ExecutionEnvironment)
		}
		if new.EncryptionKey != "" {
			create.SetEncryptionKey(new.EncryptionKey)
		}
		if new.MaxInstanceRequestConcurrency != 0 {
			create.SetMaxInstanceRequestConcurrency(new.MaxInstanceRequestConcurrency)
		}
		if new.Timeout != "" {
			create.SetTimeout(new.Timeout)
		}
		if new.ServiceAccount != "" {
			create.SetServiceAccount(new.ServiceAccount)
		}
		if new.ConditionsJSON != nil {
			create.SetConditionsJSON(new.ConditionsJSON)
		}
		if new.ObservedGeneration != 0 {
			create.SetObservedGeneration(new.ObservedGeneration)
		}
		if new.LogURI != "" {
			create.SetLogURI(new.LogURI)
		}
		if new.Etag != "" {
			create.SetEtag(new.Etag)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new Cloud Run revision history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted Cloud Run revision.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPRunRevision.Query().
		Where(
			bronzehistorygcprunrevision.ResourceID(resourceID),
			bronzehistorygcprunrevision.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current Cloud Run revision history: %w", err)
	}

	err = tx.BronzeHistoryGCPRunRevision.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close Cloud Run revision history: %w", err)
	}

	return nil
}
