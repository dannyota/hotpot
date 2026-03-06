package healthcheck

import (
	"context"
	"fmt"
	"time"

	entcompute "danny.vn/hotpot/pkg/storage/ent/gcp/compute"
	"danny.vn/hotpot/pkg/storage/ent/gcp/compute/bronzehistorygcpcomputehealthcheck"
)

// HistoryService manages health check history tracking.
type HistoryService struct {
	entClient *entcompute.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entcompute.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new health check.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entcompute.Tx, data *HealthCheckData, now time.Time) error {
	create := tx.BronzeHistoryGCPComputeHealthCheck.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
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
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entcompute.Tx, old *entcompute.BronzeGCPComputeHealthCheck, new *HealthCheckData, diff *HealthCheckDiff, now time.Time) error {
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
	create := tx.BronzeHistoryGCPComputeHealthCheck.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetProjectID(new.ProjectID)

	if new.Description != "" {
		create.SetDescription(new.Description)
	}
	if new.CreationTimestamp != "" {
		create.SetCreationTimestamp(new.CreationTimestamp)
	}
	if new.SelfLink != "" {
		create.SetSelfLink(new.SelfLink)
	}
	if new.Type != "" {
		create.SetType(new.Type)
	}
	if new.Region != "" {
		create.SetRegion(new.Region)
	}
	if new.CheckIntervalSec != 0 {
		create.SetCheckIntervalSec(new.CheckIntervalSec)
	}
	if new.TimeoutSec != 0 {
		create.SetTimeoutSec(new.TimeoutSec)
	}
	if new.HealthyThreshold != 0 {
		create.SetHealthyThreshold(new.HealthyThreshold)
	}
	if new.UnhealthyThreshold != 0 {
		create.SetUnhealthyThreshold(new.UnhealthyThreshold)
	}
	if new.TcpHealthCheckJSON != nil {
		create.SetTCPHealthCheckJSON(new.TcpHealthCheckJSON)
	}
	if new.HttpHealthCheckJSON != nil {
		create.SetHTTPHealthCheckJSON(new.HttpHealthCheckJSON)
	}
	if new.HttpsHealthCheckJSON != nil {
		create.SetHTTPSHealthCheckJSON(new.HttpsHealthCheckJSON)
	}
	if new.Http2HealthCheckJSON != nil {
		create.SetHttp2HealthCheckJSON(new.Http2HealthCheckJSON)
	}
	if new.SslHealthCheckJSON != nil {
		create.SetSslHealthCheckJSON(new.SslHealthCheckJSON)
	}
	if new.GrpcHealthCheckJSON != nil {
		create.SetGrpcHealthCheckJSON(new.GrpcHealthCheckJSON)
	}
	if new.LogConfigJSON != nil {
		create.SetLogConfigJSON(new.LogConfigJSON)
	}

	_, err = create.Save(ctx)
	return err
}

// CloseHistory closes history records for a deleted health check.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entcompute.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPComputeHealthCheck.Update().
		Where(
			bronzehistorygcpcomputehealthcheck.ResourceID(resourceID),
			bronzehistorygcpcomputehealthcheck.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if entcompute.IsNotFound(err) {
		return nil // No history to close
	}
	return err
}
