package targetpool

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputetargetpool"
)

// HistoryService manages target pool history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new target pool.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *TargetPoolData, now time.Time) error {
	create := tx.BronzeHistoryGCPComputeTargetPool.Create().
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
	if data.SessionAffinity != "" {
		create.SetSessionAffinity(data.SessionAffinity)
	}
	if data.BackupPool != "" {
		create.SetBackupPool(data.BackupPool)
	}
	if data.FailoverRatio != 0 {
		create.SetFailoverRatio(data.FailoverRatio)
	}
	if data.SecurityPolicy != "" {
		create.SetSecurityPolicy(data.SecurityPolicy)
	}
	if data.Region != "" {
		create.SetRegion(data.Region)
	}
	if data.HealthChecksJSON != nil {
		create.SetHealthChecksJSON(data.HealthChecksJSON)
	}
	if data.InstancesJSON != nil {
		create.SetInstancesJSON(data.InstancesJSON)
	}

	_, err := create.Save(ctx)
	return err
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeTargetPool, new *TargetPoolData, diff *TargetPoolDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	_, err := tx.BronzeHistoryGCPComputeTargetPool.Update().
		Where(
			bronzehistorygcpcomputetargetpool.ResourceID(old.ID),
			bronzehistorygcpcomputetargetpool.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close old history: %w", err)
	}

	// Create new history
	create := tx.BronzeHistoryGCPComputeTargetPool.Create().
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
	if new.SessionAffinity != "" {
		create.SetSessionAffinity(new.SessionAffinity)
	}
	if new.BackupPool != "" {
		create.SetBackupPool(new.BackupPool)
	}
	if new.FailoverRatio != 0 {
		create.SetFailoverRatio(new.FailoverRatio)
	}
	if new.SecurityPolicy != "" {
		create.SetSecurityPolicy(new.SecurityPolicy)
	}
	if new.Region != "" {
		create.SetRegion(new.Region)
	}
	if new.HealthChecksJSON != nil {
		create.SetHealthChecksJSON(new.HealthChecksJSON)
	}
	if new.InstancesJSON != nil {
		create.SetInstancesJSON(new.InstancesJSON)
	}

	_, err = create.Save(ctx)
	return err
}

// CloseHistory closes history records for a deleted target pool.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPComputeTargetPool.Update().
		Where(
			bronzehistorygcpcomputetargetpool.ResourceID(resourceID),
			bronzehistorygcpcomputetargetpool.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if ent.IsNotFound(err) {
		return nil
	}
	return err
}
