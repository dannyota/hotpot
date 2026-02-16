package enabledservice

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpserviceusageenabledservice"
)

// Service handles GCP Service Usage enabled service ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new enabled service ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for enabled service ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of enabled service ingestion.
type IngestResult struct {
	ProjectID      string
	ServiceCount   int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches enabled services from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	services, err := s.client.ListEnabledServices(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list enabled services: %w", err)
	}

	serviceDataList := make([]*EnabledServiceData, 0, len(services))
	for _, svc := range services {
		data, err := ConvertEnabledService(svc, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert enabled service: %w", err)
		}
		serviceDataList = append(serviceDataList, data)
	}

	if err := s.saveEnabledServices(ctx, serviceDataList); err != nil {
		return nil, fmt.Errorf("failed to save enabled services: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		ServiceCount:   len(serviceDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveEnabledServices(ctx context.Context, services []*EnabledServiceData) error {
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
		existing, err := tx.BronzeGCPServiceUsageEnabledService.Query().
			Where(bronzegcpserviceusageenabledservice.ID(serviceData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing enabled service %s: %w", serviceData.ID, err)
		}

		diff := DiffEnabledServiceData(existing, serviceData)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPServiceUsageEnabledService.UpdateOneID(serviceData.ID).
				SetCollectedAt(serviceData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for enabled service %s: %w", serviceData.ID, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeGCPServiceUsageEnabledService.Create().
				SetID(serviceData.ID).
				SetName(serviceData.Name).
				SetParent(serviceData.Parent).
				SetState(serviceData.State).
				SetProjectID(serviceData.ProjectID).
				SetCollectedAt(serviceData.CollectedAt).
				SetFirstCollectedAt(serviceData.CollectedAt)

			if serviceData.ConfigJSON != nil {
				create.SetConfigJSON(serviceData.ConfigJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create enabled service %s: %w", serviceData.ID, err)
			}
		} else {
			update := tx.BronzeGCPServiceUsageEnabledService.UpdateOneID(serviceData.ID).
				SetName(serviceData.Name).
				SetParent(serviceData.Parent).
				SetState(serviceData.State).
				SetProjectID(serviceData.ProjectID).
				SetCollectedAt(serviceData.CollectedAt)

			if serviceData.ConfigJSON != nil {
				update.SetConfigJSON(serviceData.ConfigJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update enabled service %s: %w", serviceData.ID, err)
			}
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, serviceData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for enabled service %s: %w", serviceData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, serviceData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for enabled service %s: %w", serviceData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleServices removes enabled services that were not collected in the latest run.
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

	staleServices, err := tx.BronzeGCPServiceUsageEnabledService.Query().
		Where(
			bronzegcpserviceusageenabledservice.ProjectID(projectID),
			bronzegcpserviceusageenabledservice.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, svc := range staleServices {
		if err := s.history.CloseHistory(ctx, tx, svc.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for enabled service %s: %w", svc.ID, err)
		}

		if err := tx.BronzeGCPServiceUsageEnabledService.DeleteOne(svc).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete enabled service %s: %w", svc.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
