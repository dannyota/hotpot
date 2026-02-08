package healthcheck

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzegcpcomputehealthcheck"
)

// Service handles GCP Compute health check ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new health check ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for health check ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of health check ingestion.
type IngestResult struct {
	ProjectID        string
	HealthCheckCount int
	CollectedAt      time.Time
	DurationMillis   int64
}

// Ingest fetches health checks from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch health checks from GCP
	healthChecks, err := s.client.ListHealthChecks(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list health checks: %w", err)
	}

	// Convert to data structs
	healthCheckDataList := make([]*HealthCheckData, 0, len(healthChecks))
	for _, hc := range healthChecks {
		data, err := ConvertHealthCheck(hc, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert health check: %w", err)
		}
		healthCheckDataList = append(healthCheckDataList, data)
	}

	// Save to database
	if err := s.saveHealthChecks(ctx, healthCheckDataList); err != nil {
		return nil, fmt.Errorf("failed to save health checks: %w", err)
	}

	return &IngestResult{
		ProjectID:        params.ProjectID,
		HealthCheckCount: len(healthCheckDataList),
		CollectedAt:      collectedAt,
		DurationMillis:   time.Since(startTime).Milliseconds(),
	}, nil
}

// saveHealthChecks saves health checks to the database with history tracking.
func (s *Service) saveHealthChecks(ctx context.Context, checks []*HealthCheckData) error {
	if len(checks) == 0 {
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

	for _, checkData := range checks {
		// Load existing health check
		existing, err := tx.BronzeGCPComputeHealthCheck.Query().
			Where(bronzegcpcomputehealthcheck.ID(checkData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing health check %s: %w", checkData.Name, err)
		}

		// Compute diff
		diff := DiffHealthCheckData(existing, checkData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPComputeHealthCheck.UpdateOneID(checkData.ID).
				SetCollectedAt(checkData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for health check %s: %w", checkData.Name, err)
			}
			continue
		}

		// Create or update health check
		if existing == nil {
			// Create new health check
			create := tx.BronzeGCPComputeHealthCheck.Create().
				SetID(checkData.ID).
				SetName(checkData.Name).
				SetProjectID(checkData.ProjectID).
				SetCollectedAt(checkData.CollectedAt).
				SetFirstCollectedAt(checkData.CollectedAt)

			if checkData.Description != "" {
				create.SetDescription(checkData.Description)
			}
			if checkData.CreationTimestamp != "" {
				create.SetCreationTimestamp(checkData.CreationTimestamp)
			}
			if checkData.SelfLink != "" {
				create.SetSelfLink(checkData.SelfLink)
			}
			if checkData.Type != "" {
				create.SetType(checkData.Type)
			}
			if checkData.Region != "" {
				create.SetRegion(checkData.Region)
			}
			if checkData.CheckIntervalSec != 0 {
				create.SetCheckIntervalSec(checkData.CheckIntervalSec)
			}
			if checkData.TimeoutSec != 0 {
				create.SetTimeoutSec(checkData.TimeoutSec)
			}
			if checkData.HealthyThreshold != 0 {
				create.SetHealthyThreshold(checkData.HealthyThreshold)
			}
			if checkData.UnhealthyThreshold != 0 {
				create.SetUnhealthyThreshold(checkData.UnhealthyThreshold)
			}
			if checkData.TcpHealthCheckJSON != nil {
				create.SetTCPHealthCheckJSON(checkData.TcpHealthCheckJSON)
			}
			if checkData.HttpHealthCheckJSON != nil {
				create.SetHTTPHealthCheckJSON(checkData.HttpHealthCheckJSON)
			}
			if checkData.HttpsHealthCheckJSON != nil {
				create.SetHTTPSHealthCheckJSON(checkData.HttpsHealthCheckJSON)
			}
			if checkData.Http2HealthCheckJSON != nil {
				create.SetHttp2HealthCheckJSON(checkData.Http2HealthCheckJSON)
			}
			if checkData.SslHealthCheckJSON != nil {
				create.SetSslHealthCheckJSON(checkData.SslHealthCheckJSON)
			}
			if checkData.GrpcHealthCheckJSON != nil {
				create.SetGrpcHealthCheckJSON(checkData.GrpcHealthCheckJSON)
			}
			if checkData.LogConfigJSON != nil {
				create.SetLogConfigJSON(checkData.LogConfigJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create health check %s: %w", checkData.Name, err)
			}
		} else {
			// Update existing health check
			update := tx.BronzeGCPComputeHealthCheck.UpdateOneID(checkData.ID).
				SetName(checkData.Name).
				SetProjectID(checkData.ProjectID).
				SetCollectedAt(checkData.CollectedAt)

			if checkData.Description != "" {
				update.SetDescription(checkData.Description)
			}
			if checkData.CreationTimestamp != "" {
				update.SetCreationTimestamp(checkData.CreationTimestamp)
			}
			if checkData.SelfLink != "" {
				update.SetSelfLink(checkData.SelfLink)
			}
			if checkData.Type != "" {
				update.SetType(checkData.Type)
			}
			if checkData.Region != "" {
				update.SetRegion(checkData.Region)
			}
			if checkData.CheckIntervalSec != 0 {
				update.SetCheckIntervalSec(checkData.CheckIntervalSec)
			}
			if checkData.TimeoutSec != 0 {
				update.SetTimeoutSec(checkData.TimeoutSec)
			}
			if checkData.HealthyThreshold != 0 {
				update.SetHealthyThreshold(checkData.HealthyThreshold)
			}
			if checkData.UnhealthyThreshold != 0 {
				update.SetUnhealthyThreshold(checkData.UnhealthyThreshold)
			}
			if checkData.TcpHealthCheckJSON != nil {
				update.SetTCPHealthCheckJSON(checkData.TcpHealthCheckJSON)
			}
			if checkData.HttpHealthCheckJSON != nil {
				update.SetHTTPHealthCheckJSON(checkData.HttpHealthCheckJSON)
			}
			if checkData.HttpsHealthCheckJSON != nil {
				update.SetHTTPSHealthCheckJSON(checkData.HttpsHealthCheckJSON)
			}
			if checkData.Http2HealthCheckJSON != nil {
				update.SetHttp2HealthCheckJSON(checkData.Http2HealthCheckJSON)
			}
			if checkData.SslHealthCheckJSON != nil {
				update.SetSslHealthCheckJSON(checkData.SslHealthCheckJSON)
			}
			if checkData.GrpcHealthCheckJSON != nil {
				update.SetGrpcHealthCheckJSON(checkData.GrpcHealthCheckJSON)
			}
			if checkData.LogConfigJSON != nil {
				update.SetLogConfigJSON(checkData.LogConfigJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update health check %s: %w", checkData.Name, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, checkData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for health check %s: %w", checkData.Name, err)
			}
		} else if diff.IsChanged {
			if err := s.history.UpdateHistory(ctx, tx, existing, checkData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for health check %s: %w", checkData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleHealthChecks removes health checks that were not collected in the latest run.
// Also closes history records for deleted health checks.
func (s *Service) DeleteStaleHealthChecks(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale health checks
	staleChecks, err := tx.BronzeGCPComputeHealthCheck.Query().
		Where(
			bronzegcpcomputehealthcheck.ProjectID(projectID),
			bronzegcpcomputehealthcheck.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale health check
	for _, check := range staleChecks {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, check.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for health check %s: %w", check.ID, err)
		}

		// Delete health check
		if err := tx.BronzeGCPComputeHealthCheck.DeleteOne(check).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete health check %s: %w", check.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
