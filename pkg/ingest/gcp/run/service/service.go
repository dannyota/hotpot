package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcprunservice"
)

// Service handles Cloud Run service ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new Cloud Run service ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for Cloud Run service ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of Cloud Run service ingestion.
type IngestResult struct {
	ProjectID      string
	ServiceCount   int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches Cloud Run services from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch services from GCP
	rawServices, err := s.client.ListServices(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list Cloud Run services: %w", err)
	}

	// Convert to service data
	serviceDataList := make([]*ServiceData, 0, len(rawServices))
	for _, svc := range rawServices {
		data, err := ConvertService(svc, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert Cloud Run service: %w", err)
		}
		serviceDataList = append(serviceDataList, data)
	}

	// Save to database
	if err := s.saveServices(ctx, serviceDataList); err != nil {
		return nil, fmt.Errorf("failed to save Cloud Run services: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		ServiceCount:   len(serviceDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveServices saves Cloud Run services to the database with history tracking.
func (s *Service) saveServices(ctx context.Context, services []*ServiceData) error {
	if len(services) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, serviceData := range services {
		// Load existing service
		existing, err := tx.BronzeGCPRunService.Query().
			Where(bronzegcprunservice.ID(serviceData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing Cloud Run service %s: %w", serviceData.ID, err)
		}

		// Compute diff
		diff := DiffServiceData(existing, serviceData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPRunService.UpdateOneID(serviceData.ID).
				SetCollectedAt(serviceData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for Cloud Run service %s: %w", serviceData.ID, err)
			}
			continue
		}

		// Create or update service
		if existing == nil {
			create := tx.BronzeGCPRunService.Create().
				SetID(serviceData.ID).
				SetName(serviceData.Name).
				SetProjectID(serviceData.ProjectID).
				SetLocation(serviceData.Location).
				SetReconciling(serviceData.Reconciling).
				SetCollectedAt(serviceData.CollectedAt).
				SetFirstCollectedAt(serviceData.CollectedAt)

			if serviceData.Description != "" {
				create.SetDescription(serviceData.Description)
			}
			if serviceData.UID != "" {
				create.SetUID(serviceData.UID)
			}
			if serviceData.Generation != 0 {
				create.SetGeneration(serviceData.Generation)
			}
			if serviceData.LabelsJSON != nil {
				create.SetLabelsJSON(serviceData.LabelsJSON)
			}
			if serviceData.AnnotationsJSON != nil {
				create.SetAnnotationsJSON(serviceData.AnnotationsJSON)
			}
			if serviceData.CreateTime != "" {
				create.SetCreateTime(serviceData.CreateTime)
			}
			if serviceData.UpdateTime != "" {
				create.SetUpdateTime(serviceData.UpdateTime)
			}
			if serviceData.DeleteTime != "" {
				create.SetDeleteTime(serviceData.DeleteTime)
			}
			if serviceData.Creator != "" {
				create.SetCreator(serviceData.Creator)
			}
			if serviceData.LastModifier != "" {
				create.SetLastModifier(serviceData.LastModifier)
			}
			if serviceData.Ingress != 0 {
				create.SetIngress(serviceData.Ingress)
			}
			if serviceData.LaunchStage != 0 {
				create.SetLaunchStage(serviceData.LaunchStage)
			}
			if serviceData.TemplateJSON != nil {
				create.SetTemplateJSON(serviceData.TemplateJSON)
			}
			if serviceData.TrafficJSON != nil {
				create.SetTrafficJSON(serviceData.TrafficJSON)
			}
			if serviceData.URI != "" {
				create.SetURI(serviceData.URI)
			}
			if serviceData.ObservedGeneration != 0 {
				create.SetObservedGeneration(serviceData.ObservedGeneration)
			}
			if serviceData.TerminalConditionJSON != nil {
				create.SetTerminalConditionJSON(serviceData.TerminalConditionJSON)
			}
			if serviceData.ConditionsJSON != nil {
				create.SetConditionsJSON(serviceData.ConditionsJSON)
			}
			if serviceData.LatestReadyRevision != "" {
				create.SetLatestReadyRevision(serviceData.LatestReadyRevision)
			}
			if serviceData.LatestCreatedRevision != "" {
				create.SetLatestCreatedRevision(serviceData.LatestCreatedRevision)
			}
			if serviceData.TrafficStatusesJSON != nil {
				create.SetTrafficStatusesJSON(serviceData.TrafficStatusesJSON)
			}
			if serviceData.Etag != "" {
				create.SetEtag(serviceData.Etag)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create Cloud Run service %s: %w", serviceData.ID, err)
			}
		} else {
			update := tx.BronzeGCPRunService.UpdateOneID(serviceData.ID).
				SetName(serviceData.Name).
				SetProjectID(serviceData.ProjectID).
				SetLocation(serviceData.Location).
				SetReconciling(serviceData.Reconciling).
				SetCollectedAt(serviceData.CollectedAt)

			if serviceData.Description != "" {
				update.SetDescription(serviceData.Description)
			}
			if serviceData.UID != "" {
				update.SetUID(serviceData.UID)
			}
			if serviceData.Generation != 0 {
				update.SetGeneration(serviceData.Generation)
			}
			if serviceData.LabelsJSON != nil {
				update.SetLabelsJSON(serviceData.LabelsJSON)
			}
			if serviceData.AnnotationsJSON != nil {
				update.SetAnnotationsJSON(serviceData.AnnotationsJSON)
			}
			if serviceData.CreateTime != "" {
				update.SetCreateTime(serviceData.CreateTime)
			}
			if serviceData.UpdateTime != "" {
				update.SetUpdateTime(serviceData.UpdateTime)
			}
			if serviceData.DeleteTime != "" {
				update.SetDeleteTime(serviceData.DeleteTime)
			}
			if serviceData.Creator != "" {
				update.SetCreator(serviceData.Creator)
			}
			if serviceData.LastModifier != "" {
				update.SetLastModifier(serviceData.LastModifier)
			}
			if serviceData.Ingress != 0 {
				update.SetIngress(serviceData.Ingress)
			}
			if serviceData.LaunchStage != 0 {
				update.SetLaunchStage(serviceData.LaunchStage)
			}
			if serviceData.TemplateJSON != nil {
				update.SetTemplateJSON(serviceData.TemplateJSON)
			}
			if serviceData.TrafficJSON != nil {
				update.SetTrafficJSON(serviceData.TrafficJSON)
			}
			if serviceData.URI != "" {
				update.SetURI(serviceData.URI)
			}
			if serviceData.ObservedGeneration != 0 {
				update.SetObservedGeneration(serviceData.ObservedGeneration)
			}
			if serviceData.TerminalConditionJSON != nil {
				update.SetTerminalConditionJSON(serviceData.TerminalConditionJSON)
			}
			if serviceData.ConditionsJSON != nil {
				update.SetConditionsJSON(serviceData.ConditionsJSON)
			}
			if serviceData.LatestReadyRevision != "" {
				update.SetLatestReadyRevision(serviceData.LatestReadyRevision)
			}
			if serviceData.LatestCreatedRevision != "" {
				update.SetLatestCreatedRevision(serviceData.LatestCreatedRevision)
			}
			if serviceData.TrafficStatusesJSON != nil {
				update.SetTrafficStatusesJSON(serviceData.TrafficStatusesJSON)
			}
			if serviceData.Etag != "" {
				update.SetEtag(serviceData.Etag)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update Cloud Run service %s: %w", serviceData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, serviceData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for Cloud Run service %s: %w", serviceData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, serviceData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for Cloud Run service %s: %w", serviceData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleServices removes services that were not collected in the latest run.
func (s *Service) DeleteStaleServices(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	staleServices, err := tx.BronzeGCPRunService.Query().
		Where(
			bronzegcprunservice.ProjectID(projectID),
			bronzegcprunservice.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, svc := range staleServices {
		if err := s.history.CloseHistory(ctx, tx, svc.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for Cloud Run service %s: %w", svc.ID, err)
		}

		if err := tx.BronzeGCPRunService.DeleteOne(svc).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete Cloud Run service %s: %w", svc.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
