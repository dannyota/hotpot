package sslpolicy

import (
	"context"
	"fmt"
	"time"

	entcompute "danny.vn/hotpot/pkg/storage/ent/gcp/compute"
	"danny.vn/hotpot/pkg/storage/ent/gcp/compute/bronzehistorygcpcomputesslpolicy"
)

// HistoryService manages SSL policy history tracking.
type HistoryService struct {
	entClient *entcompute.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entcompute.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new SSL policy.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entcompute.Tx, data *SslPolicyData, now time.Time) error {
	create := tx.BronzeHistoryGCPComputeSslPolicy.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetProjectID(data.ProjectID)

	if data.Description != "" {
		create.SetDescription(data.Description)
	}
	if data.SelfLink != "" {
		create.SetSelfLink(data.SelfLink)
	}
	if data.CreationTimestamp != "" {
		create.SetCreationTimestamp(data.CreationTimestamp)
	}
	if data.Profile != "" {
		create.SetProfile(data.Profile)
	}
	if data.MinTlsVersion != "" {
		create.SetMinTLSVersion(data.MinTlsVersion)
	}
	if data.Fingerprint != "" {
		create.SetFingerprint(data.Fingerprint)
	}
	if data.CustomFeaturesJSON != nil {
		create.SetCustomFeaturesJSON(data.CustomFeaturesJSON)
	}
	if data.EnabledFeaturesJSON != nil {
		create.SetEnabledFeaturesJSON(data.EnabledFeaturesJSON)
	}
	if data.WarningsJSON != nil {
		create.SetWarningsJSON(data.WarningsJSON)
	}

	_, err := create.Save(ctx)
	return err
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entcompute.Tx, old *entcompute.BronzeGCPComputeSslPolicy, new *SslPolicyData, diff *SslPolicyDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	_, err := tx.BronzeHistoryGCPComputeSslPolicy.Update().
		Where(
			bronzehistorygcpcomputesslpolicy.ResourceID(old.ID),
			bronzehistorygcpcomputesslpolicy.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close old history: %w", err)
	}

	// Create new history
	create := tx.BronzeHistoryGCPComputeSslPolicy.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetProjectID(new.ProjectID)

	if new.Description != "" {
		create.SetDescription(new.Description)
	}
	if new.SelfLink != "" {
		create.SetSelfLink(new.SelfLink)
	}
	if new.CreationTimestamp != "" {
		create.SetCreationTimestamp(new.CreationTimestamp)
	}
	if new.Profile != "" {
		create.SetProfile(new.Profile)
	}
	if new.MinTlsVersion != "" {
		create.SetMinTLSVersion(new.MinTlsVersion)
	}
	if new.Fingerprint != "" {
		create.SetFingerprint(new.Fingerprint)
	}
	if new.CustomFeaturesJSON != nil {
		create.SetCustomFeaturesJSON(new.CustomFeaturesJSON)
	}
	if new.EnabledFeaturesJSON != nil {
		create.SetEnabledFeaturesJSON(new.EnabledFeaturesJSON)
	}
	if new.WarningsJSON != nil {
		create.SetWarningsJSON(new.WarningsJSON)
	}

	_, err = create.Save(ctx)
	return err
}

// CloseHistory closes history records for a deleted SSL policy.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entcompute.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPComputeSslPolicy.Update().
		Where(
			bronzehistorygcpcomputesslpolicy.ResourceID(resourceID),
			bronzehistorygcpcomputesslpolicy.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if entcompute.IsNotFound(err) {
		return nil // No history to close
	}
	return err
}
