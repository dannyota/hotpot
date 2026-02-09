package targetinstance

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcomputetargetinstance"
)

// HistoryService manages target instance history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new target instance.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *TargetInstanceData, now time.Time) error {
	create := tx.BronzeHistoryGCPComputeTargetInstance.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetProjectID(data.ProjectID)

	if data.Description != "" {
		create.SetDescription(data.Description)
	}
	if data.Zone != "" {
		create.SetZone(data.Zone)
	}
	if data.Instance != "" {
		create.SetInstance(data.Instance)
	}
	if data.Network != "" {
		create.SetNetwork(data.Network)
	}
	if data.NatPolicy != "" {
		create.SetNatPolicy(data.NatPolicy)
	}
	if data.SecurityPolicy != "" {
		create.SetSecurityPolicy(data.SecurityPolicy)
	}
	if data.SelfLink != "" {
		create.SetSelfLink(data.SelfLink)
	}
	if data.CreationTimestamp != "" {
		create.SetCreationTimestamp(data.CreationTimestamp)
	}

	_, err := create.Save(ctx)
	return err
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeTargetInstance, new *TargetInstanceData, diff *TargetInstanceDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	_, err := tx.BronzeHistoryGCPComputeTargetInstance.Update().
		Where(
			bronzehistorygcpcomputetargetinstance.ResourceID(old.ID),
			bronzehistorygcpcomputetargetinstance.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close old history: %w", err)
	}

	// Create new history
	create := tx.BronzeHistoryGCPComputeTargetInstance.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetProjectID(new.ProjectID)

	if new.Description != "" {
		create.SetDescription(new.Description)
	}
	if new.Zone != "" {
		create.SetZone(new.Zone)
	}
	if new.Instance != "" {
		create.SetInstance(new.Instance)
	}
	if new.Network != "" {
		create.SetNetwork(new.Network)
	}
	if new.NatPolicy != "" {
		create.SetNatPolicy(new.NatPolicy)
	}
	if new.SecurityPolicy != "" {
		create.SetSecurityPolicy(new.SecurityPolicy)
	}
	if new.SelfLink != "" {
		create.SetSelfLink(new.SelfLink)
	}
	if new.CreationTimestamp != "" {
		create.SetCreationTimestamp(new.CreationTimestamp)
	}

	_, err = create.Save(ctx)
	return err
}

// CloseHistory closes history records for a deleted target instance.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPComputeTargetInstance.Update().
		Where(
			bronzehistorygcpcomputetargetinstance.ResourceID(resourceID),
			bronzehistorygcpcomputetargetinstance.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if ent.IsNotFound(err) {
		return nil // No history to close
	}
	return err
}
