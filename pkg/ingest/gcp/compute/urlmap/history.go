package urlmap

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputeurlmap"
)

// HistoryService manages URL map history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new URL map.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *UrlMapData, now time.Time) error {
	create := tx.BronzeHistoryGCPComputeUrlMap.Create().
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
	if data.DefaultService != "" {
		create.SetDefaultService(data.DefaultService)
	}
	if data.Region != "" {
		create.SetRegion(data.Region)
	}
	if data.HostRulesJSON != nil {
		create.SetHostRulesJSON(data.HostRulesJSON)
	}
	if data.PathMatchersJSON != nil {
		create.SetPathMatchersJSON(data.PathMatchersJSON)
	}
	if data.TestsJSON != nil {
		create.SetTestsJSON(data.TestsJSON)
	}
	if data.DefaultRouteActionJSON != nil {
		create.SetDefaultRouteActionJSON(data.DefaultRouteActionJSON)
	}
	if data.DefaultUrlRedirectJSON != nil {
		create.SetDefaultURLRedirectJSON(data.DefaultUrlRedirectJSON)
	}
	if data.HeaderActionJSON != nil {
		create.SetHeaderActionJSON(data.HeaderActionJSON)
	}

	_, err := create.Save(ctx)
	return err
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeUrlMap, new *UrlMapData, diff *UrlMapDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	_, err := tx.BronzeHistoryGCPComputeUrlMap.Update().
		Where(
			bronzehistorygcpcomputeurlmap.ResourceID(old.ID),
			bronzehistorygcpcomputeurlmap.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close old history: %w", err)
	}

	// Create new history
	create := tx.BronzeHistoryGCPComputeUrlMap.Create().
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
	if new.DefaultService != "" {
		create.SetDefaultService(new.DefaultService)
	}
	if new.Region != "" {
		create.SetRegion(new.Region)
	}
	if new.HostRulesJSON != nil {
		create.SetHostRulesJSON(new.HostRulesJSON)
	}
	if new.PathMatchersJSON != nil {
		create.SetPathMatchersJSON(new.PathMatchersJSON)
	}
	if new.TestsJSON != nil {
		create.SetTestsJSON(new.TestsJSON)
	}
	if new.DefaultRouteActionJSON != nil {
		create.SetDefaultRouteActionJSON(new.DefaultRouteActionJSON)
	}
	if new.DefaultUrlRedirectJSON != nil {
		create.SetDefaultURLRedirectJSON(new.DefaultUrlRedirectJSON)
	}
	if new.HeaderActionJSON != nil {
		create.SetHeaderActionJSON(new.HeaderActionJSON)
	}

	_, err = create.Save(ctx)
	return err
}

// CloseHistory closes history records for a deleted URL map.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPComputeUrlMap.Update().
		Where(
			bronzehistorygcpcomputeurlmap.ResourceID(resourceID),
			bronzehistorygcpcomputeurlmap.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if ent.IsNotFound(err) {
		return nil
	}
	return err
}
