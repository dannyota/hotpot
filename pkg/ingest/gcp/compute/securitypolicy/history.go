package securitypolicy

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcomputesecuritypolicy"
)

// HistoryService manages security policy history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new security policy.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *SecurityPolicyData, now time.Time) error {
	create := tx.BronzeHistoryGCPComputeSecurityPolicy.Create().
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
	if data.Fingerprint != "" {
		create.SetFingerprint(data.Fingerprint)
	}
	if data.RulesJSON != nil {
		create.SetRulesJSON(data.RulesJSON)
	}
	if data.AssociationsJSON != nil {
		create.SetAssociationsJSON(data.AssociationsJSON)
	}
	if data.AdaptiveProtectionConfigJSON != nil {
		create.SetAdaptiveProtectionConfigJSON(data.AdaptiveProtectionConfigJSON)
	}
	if data.AdvancedOptionsConfigJSON != nil {
		create.SetAdvancedOptionsConfigJSON(data.AdvancedOptionsConfigJSON)
	}
	if data.DdosProtectionConfigJSON != nil {
		create.SetDdosProtectionConfigJSON(data.DdosProtectionConfigJSON)
	}
	if data.RecaptchaOptionsConfigJSON != nil {
		create.SetRecaptchaOptionsConfigJSON(data.RecaptchaOptionsConfigJSON)
	}
	if data.LabelsJSON != nil {
		create.SetLabelsJSON(data.LabelsJSON)
	}

	_, err := create.Save(ctx)
	return err
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeSecurityPolicy, new *SecurityPolicyData, diff *SecurityPolicyDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	_, err := tx.BronzeHistoryGCPComputeSecurityPolicy.Update().
		Where(
			bronzehistorygcpcomputesecuritypolicy.ResourceID(old.ID),
			bronzehistorygcpcomputesecuritypolicy.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close old history: %w", err)
	}

	// Create new history
	create := tx.BronzeHistoryGCPComputeSecurityPolicy.Create().
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
	if new.Fingerprint != "" {
		create.SetFingerprint(new.Fingerprint)
	}
	if new.RulesJSON != nil {
		create.SetRulesJSON(new.RulesJSON)
	}
	if new.AssociationsJSON != nil {
		create.SetAssociationsJSON(new.AssociationsJSON)
	}
	if new.AdaptiveProtectionConfigJSON != nil {
		create.SetAdaptiveProtectionConfigJSON(new.AdaptiveProtectionConfigJSON)
	}
	if new.AdvancedOptionsConfigJSON != nil {
		create.SetAdvancedOptionsConfigJSON(new.AdvancedOptionsConfigJSON)
	}
	if new.DdosProtectionConfigJSON != nil {
		create.SetDdosProtectionConfigJSON(new.DdosProtectionConfigJSON)
	}
	if new.RecaptchaOptionsConfigJSON != nil {
		create.SetRecaptchaOptionsConfigJSON(new.RecaptchaOptionsConfigJSON)
	}
	if new.LabelsJSON != nil {
		create.SetLabelsJSON(new.LabelsJSON)
	}

	_, err = create.Save(ctx)
	return err
}

// CloseHistory closes history records for a deleted security policy.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPComputeSecurityPolicy.Update().
		Where(
			bronzehistorygcpcomputesecuritypolicy.ResourceID(resourceID),
			bronzehistorygcpcomputesecuritypolicy.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if ent.IsNotFound(err) {
		return nil // No history to close
	}
	return err
}
