package targethttpproxy

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputetargethttpproxy"
)

// HistoryService manages target HTTP proxy history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new target HTTP proxy.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *TargetHttpProxyData, now time.Time) error {
	create := tx.BronzeHistoryGCPComputeTargetHttpProxy.Create().
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
	if data.Fingerprint != "" {
		create.SetFingerprint(data.Fingerprint)
	}
	if data.UrlMap != "" {
		create.SetURLMap(data.UrlMap)
	}
	if data.ProxyBind {
		create.SetProxyBind(data.ProxyBind)
	}
	if data.HttpKeepAliveTimeoutSec != 0 {
		create.SetHTTPKeepAliveTimeoutSec(data.HttpKeepAliveTimeoutSec)
	}
	if data.Region != "" {
		create.SetRegion(data.Region)
	}

	_, err := create.Save(ctx)
	return err
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeTargetHttpProxy, new *TargetHttpProxyData, diff *TargetHttpProxyDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	_, err := tx.BronzeHistoryGCPComputeTargetHttpProxy.Update().
		Where(
			bronzehistorygcpcomputetargethttpproxy.ResourceID(old.ID),
			bronzehistorygcpcomputetargethttpproxy.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close old history: %w", err)
	}

	// Create new history
	create := tx.BronzeHistoryGCPComputeTargetHttpProxy.Create().
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
	if new.Fingerprint != "" {
		create.SetFingerprint(new.Fingerprint)
	}
	if new.UrlMap != "" {
		create.SetURLMap(new.UrlMap)
	}
	if new.ProxyBind {
		create.SetProxyBind(new.ProxyBind)
	}
	if new.HttpKeepAliveTimeoutSec != 0 {
		create.SetHTTPKeepAliveTimeoutSec(new.HttpKeepAliveTimeoutSec)
	}
	if new.Region != "" {
		create.SetRegion(new.Region)
	}

	_, err = create.Save(ctx)
	return err
}

// CloseHistory closes history records for a deleted target HTTP proxy.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPComputeTargetHttpProxy.Update().
		Where(
			bronzehistorygcpcomputetargethttpproxy.ResourceID(resourceID),
			bronzehistorygcpcomputetargethttpproxy.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if ent.IsNotFound(err) {
		return nil
	}
	return err
}
