package targettcpproxy

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputetargettcpproxy"
)

// HistoryService manages target TCP proxy history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new target TCP proxy.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *TargetTcpProxyData, now time.Time) error {
	create := tx.BronzeHistoryGCPComputeTargetTcpProxy.Create().
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
	if data.Service != "" {
		create.SetService(data.Service)
	}
	if data.ProxyBind {
		create.SetProxyBind(data.ProxyBind)
	}
	if data.ProxyHeader != "" {
		create.SetProxyHeader(data.ProxyHeader)
	}
	if data.Region != "" {
		create.SetRegion(data.Region)
	}

	_, err := create.Save(ctx)
	return err
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeTargetTcpProxy, new *TargetTcpProxyData, diff *TargetTcpProxyDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	_, err := tx.BronzeHistoryGCPComputeTargetTcpProxy.Update().
		Where(
			bronzehistorygcpcomputetargettcpproxy.ResourceID(old.ID),
			bronzehistorygcpcomputetargettcpproxy.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close old history: %w", err)
	}

	// Create new history
	create := tx.BronzeHistoryGCPComputeTargetTcpProxy.Create().
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
	if new.Service != "" {
		create.SetService(new.Service)
	}
	if new.ProxyBind {
		create.SetProxyBind(new.ProxyBind)
	}
	if new.ProxyHeader != "" {
		create.SetProxyHeader(new.ProxyHeader)
	}
	if new.Region != "" {
		create.SetRegion(new.Region)
	}

	_, err = create.Save(ctx)
	return err
}

// CloseHistory closes history records for a deleted target TCP proxy.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPComputeTargetTcpProxy.Update().
		Where(
			bronzehistorygcpcomputetargettcpproxy.ResourceID(resourceID),
			bronzehistorygcpcomputetargettcpproxy.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if ent.IsNotFound(err) {
		return nil
	}
	return err
}
