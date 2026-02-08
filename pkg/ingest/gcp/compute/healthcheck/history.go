package healthcheck

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputehealthcheck"
)

// HistoryService manages health check history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new health check.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *HealthCheckData, now time.Time) error {
	create := tx.BronzeHistoryGCPComputeHealthCheck.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetProjectID(data.ProjectID)

	if data.Description != "" {
		create.SetDescription(data.Description)
	}
	if data.CreationTimestamp != "" {
		create.SetCreationTimestamp(data.CreationTimestamp)
	}
	if data.SelfLink != "" {
		create.SetSelfLink(data.SelfLink)
	}
	if data.Type != "" {
		create.SetType(data.Type)
	}
	if data.Region != "" {
		create.SetRegion(data.Region)
	}
	if data.CheckIntervalSec != 0 {
		create.SetCheckIntervalSec(data.CheckIntervalSec)
	}
	if data.TimeoutSec != 0 {
		create.SetTimeoutSec(data.TimeoutSec)
	}
	if data.HealthyThreshold != 0 {
		create.SetHealthyThreshold(data.HealthyThreshold)
	}
	if data.UnhealthyThreshold != 0 {
		create.SetUnhealthyThreshold(data.UnhealthyThreshold)
	}
	if data.TcpHealthCheckJSON != nil {
		create.SetTCPHealthCheckJSON(data.TcpHealthCheckJSON)
	}
	if data.HttpHealthCheckJSON != nil {
		create.SetHTTPHealthCheckJSON(data.HttpHealthCheckJSON)
	}
	if data.HttpsHealthCheckJSON != nil {
		create.SetHTTPSHealthCheckJSON(data.HttpsHealthCheckJSON)
	}
	if data.Http2HealthCheckJSON != nil {
		create.SetHttp2HealthCheckJSON(data.Http2HealthCheckJSON)
	}
	if data.SslHealthCheckJSON != nil {
		create.SetSslHealthCheckJSON(data.SslHealthCheckJSON)
	}
	if data.GrpcHealthCheckJSON != nil {
		create.SetGrpcHealthCheckJSON(data.GrpcHealthCheckJSON)
	}
	if data.LogConfigJSON != nil {
		create.SetLogConfigJSON(data.LogConfigJSON)
	}

	_, err := create.Save(ctx)
	return err
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeHealthCheck, new *HealthCheckData, diff *HealthCheckDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	_, err := tx.BronzeHistoryGCPComputeHealthCheck.Update().
		Where(
			bronzehistorygcpcomputehealthcheck.ResourceID(old.ID),
			bronzehistorygcpcomputehealthcheck.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close old history: %w", err)
	}

	// Create new history
	return h.CreateHistory(ctx, tx, new, now)
}

// CloseHistory closes history records for a deleted health check.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPComputeHealthCheck.Update().
		Where(
			bronzehistorygcpcomputehealthcheck.ResourceID(resourceID),
			bronzehistorygcpcomputehealthcheck.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if ent.IsNotFound(err) {
		return nil // No history to close
	}
	return err
}
