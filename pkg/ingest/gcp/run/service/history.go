package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcprunservice"
)

// HistoryService manages Cloud Run service history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new Cloud Run service.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *ServiceData, now time.Time) error {
	create := tx.BronzeHistoryGCPRunService.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetProjectID(data.ProjectID).
		SetLocation(data.Location).
		SetReconciling(data.Reconciling)

	if data.Description != "" {
		create.SetDescription(data.Description)
	}
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
	if data.Creator != "" {
		create.SetCreator(data.Creator)
	}
	if data.LastModifier != "" {
		create.SetLastModifier(data.LastModifier)
	}
	if data.Ingress != 0 {
		create.SetIngress(data.Ingress)
	}
	if data.LaunchStage != 0 {
		create.SetLaunchStage(data.LaunchStage)
	}
	if data.TemplateJSON != nil {
		create.SetTemplateJSON(data.TemplateJSON)
	}
	if data.TrafficJSON != nil {
		create.SetTrafficJSON(data.TrafficJSON)
	}
	if data.URI != "" {
		create.SetURI(data.URI)
	}
	if data.ObservedGeneration != 0 {
		create.SetObservedGeneration(data.ObservedGeneration)
	}
	if data.TerminalConditionJSON != nil {
		create.SetTerminalConditionJSON(data.TerminalConditionJSON)
	}
	if data.ConditionsJSON != nil {
		create.SetConditionsJSON(data.ConditionsJSON)
	}
	if data.LatestReadyRevision != "" {
		create.SetLatestReadyRevision(data.LatestReadyRevision)
	}
	if data.LatestCreatedRevision != "" {
		create.SetLatestCreatedRevision(data.LatestCreatedRevision)
	}
	if data.TrafficStatusesJSON != nil {
		create.SetTrafficStatusesJSON(data.TrafficStatusesJSON)
	}
	if data.Etag != "" {
		create.SetEtag(data.Etag)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Cloud Run service history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed Cloud Run service.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPRunService, new *ServiceData, diff *ServiceDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPRunService.Query().
		Where(
			bronzehistorygcprunservice.ResourceID(old.ID),
			bronzehistorygcprunservice.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current Cloud Run service history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPRunService.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current Cloud Run service history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPRunService.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetProjectID(new.ProjectID).
			SetLocation(new.Location).
			SetReconciling(new.Reconciling)

		if new.Description != "" {
			create.SetDescription(new.Description)
		}
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
		if new.Creator != "" {
			create.SetCreator(new.Creator)
		}
		if new.LastModifier != "" {
			create.SetLastModifier(new.LastModifier)
		}
		if new.Ingress != 0 {
			create.SetIngress(new.Ingress)
		}
		if new.LaunchStage != 0 {
			create.SetLaunchStage(new.LaunchStage)
		}
		if new.TemplateJSON != nil {
			create.SetTemplateJSON(new.TemplateJSON)
		}
		if new.TrafficJSON != nil {
			create.SetTrafficJSON(new.TrafficJSON)
		}
		if new.URI != "" {
			create.SetURI(new.URI)
		}
		if new.ObservedGeneration != 0 {
			create.SetObservedGeneration(new.ObservedGeneration)
		}
		if new.TerminalConditionJSON != nil {
			create.SetTerminalConditionJSON(new.TerminalConditionJSON)
		}
		if new.ConditionsJSON != nil {
			create.SetConditionsJSON(new.ConditionsJSON)
		}
		if new.LatestReadyRevision != "" {
			create.SetLatestReadyRevision(new.LatestReadyRevision)
		}
		if new.LatestCreatedRevision != "" {
			create.SetLatestCreatedRevision(new.LatestCreatedRevision)
		}
		if new.TrafficStatusesJSON != nil {
			create.SetTrafficStatusesJSON(new.TrafficStatusesJSON)
		}
		if new.Etag != "" {
			create.SetEtag(new.Etag)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new Cloud Run service history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted Cloud Run service.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPRunService.Query().
		Where(
			bronzehistorygcprunservice.ResourceID(resourceID),
			bronzehistorygcprunservice.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current Cloud Run service history: %w", err)
	}

	err = tx.BronzeHistoryGCPRunService.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close Cloud Run service history: %w", err)
	}

	return nil
}
