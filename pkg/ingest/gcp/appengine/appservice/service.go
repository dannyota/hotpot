package appservice

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpappengineapplication"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpappengineservice"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpappengineapplication"
)

// Service handles App Engine service ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new App Engine service ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for service ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of service ingestion.
type IngestResult struct {
	ProjectID      string
	ServiceCount   int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches App Engine services from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch services from GCP
	services, err := s.client.ListServices(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	// Convert to data structs
	serviceDataList := make([]*ServiceData, 0, len(services))
	for _, svc := range services {
		data, err := ConvertService(svc, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert service: %w", err)
		}
		serviceDataList = append(serviceDataList, data)
	}

	// Save to database
	if err := s.saveServices(ctx, serviceDataList, params.ProjectID); err != nil {
		return nil, fmt.Errorf("failed to save services: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		ServiceCount:   len(serviceDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveServices saves App Engine services to the database with history tracking.
func (s *Service) saveServices(ctx context.Context, services []*ServiceData, projectID string) error {
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

	// Look up the parent application for edge linking
	parentApp, err := tx.BronzeGCPAppEngineApplication.Query().
		Where(bronzegcpappengineapplication.ProjectID(projectID)).
		First(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to find parent application for project %s: %w", projectID, err)
	}

	// Look up the current application history ID for service history records
	appHistory, err := tx.BronzeHistoryGCPAppEngineApplication.Query().
		Where(
			bronzehistorygcpappengineapplication.ResourceID(parentApp.ID),
			bronzehistorygcpappengineapplication.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to find application history for %s: %w", parentApp.ID, err)
	}

	for _, svcData := range services {
		// Load existing service
		existing, err := tx.BronzeGCPAppEngineService.Query().
			Where(bronzegcpappengineservice.ID(svcData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing service %s: %w", svcData.ID, err)
		}

		// Compute diff
		diff := DiffServiceData(existing, svcData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPAppEngineService.UpdateOneID(svcData.ID).
				SetCollectedAt(svcData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for service %s: %w", svcData.ID, err)
			}
			continue
		}

		// Create or update service
		if existing == nil {
			create := tx.BronzeGCPAppEngineService.Create().
				SetID(svcData.ID).
				SetName(svcData.Name).
				SetProjectID(svcData.ProjectID).
				SetCollectedAt(svcData.CollectedAt).
				SetFirstCollectedAt(svcData.CollectedAt).
				SetApplication(parentApp)

			if svcData.SplitJSON != nil {
				create.SetSplitJSON(svcData.SplitJSON)
			}
			if svcData.LabelsJSON != nil {
				create.SetLabelsJSON(svcData.LabelsJSON)
			}
			if svcData.NetworkSettingsJSON != nil {
				create.SetNetworkSettingsJSON(svcData.NetworkSettingsJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create service %s: %w", svcData.ID, err)
			}
		} else {
			update := tx.BronzeGCPAppEngineService.UpdateOneID(svcData.ID).
				SetName(svcData.Name).
				SetProjectID(svcData.ProjectID).
				SetCollectedAt(svcData.CollectedAt)

			if svcData.SplitJSON != nil {
				update.SetSplitJSON(svcData.SplitJSON)
			}
			if svcData.LabelsJSON != nil {
				update.SetLabelsJSON(svcData.LabelsJSON)
			}
			if svcData.NetworkSettingsJSON != nil {
				update.SetNetworkSettingsJSON(svcData.NetworkSettingsJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update service %s: %w", svcData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, svcData, appHistory.HistoryID, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for service %s: %w", svcData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, svcData, diff, appHistory.HistoryID, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for service %s: %w", svcData.ID, err)
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

	staleServices, err := tx.BronzeGCPAppEngineService.Query().
		Where(
			bronzegcpappengineservice.ProjectID(projectID),
			bronzegcpappengineservice.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to find stale services: %w", err)
	}

	for _, svc := range staleServices {
		if err := s.history.CloseHistory(ctx, tx, svc.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for service %s: %w", svc.ID, err)
		}

		if err := tx.BronzeGCPAppEngineService.DeleteOne(svc).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete service %s: %w", svc.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
