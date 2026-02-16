package dnspolicy

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpdnspolicy"
)

// HistoryService handles history tracking for DNS policies.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new DNS policy.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, policyData *PolicyData, now time.Time) error {
	create := tx.BronzeHistoryGCPDNSPolicy.Create().
		SetResourceID(policyData.ID).
		SetValidFrom(now).
		SetCollectedAt(policyData.CollectedAt).
		SetFirstCollectedAt(policyData.CollectedAt).
		SetName(policyData.Name).
		SetEnableInboundForwarding(policyData.EnableInboundForwarding).
		SetEnableLogging(policyData.EnableLogging).
		SetProjectID(policyData.ProjectID)

	if policyData.Description != "" {
		create.SetDescription(policyData.Description)
	}
	if policyData.NetworksJSON != nil {
		create.SetNetworksJSON(policyData.NetworksJSON)
	}
	if policyData.AlternativeNameServerConfigJSON != nil {
		create.SetAlternativeNameServerConfigJSON(policyData.AlternativeNameServerConfigJSON)
	}

	_, err := create.Save(ctx)
	return err
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPDNSPolicy, new *PolicyData, diff *PolicyDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	_, err := tx.BronzeHistoryGCPDNSPolicy.Update().
		Where(
			bronzehistorygcpdnspolicy.ResourceID(old.ID),
			bronzehistorygcpdnspolicy.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close old history: %w", err)
	}

	// Create new history
	create := tx.BronzeHistoryGCPDNSPolicy.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetEnableInboundForwarding(new.EnableInboundForwarding).
		SetEnableLogging(new.EnableLogging).
		SetProjectID(new.ProjectID)

	if new.Description != "" {
		create.SetDescription(new.Description)
	}
	if new.NetworksJSON != nil {
		create.SetNetworksJSON(new.NetworksJSON)
	}
	if new.AlternativeNameServerConfigJSON != nil {
		create.SetAlternativeNameServerConfigJSON(new.AlternativeNameServerConfigJSON)
	}

	_, err = create.Save(ctx)
	return err
}

// CloseHistory closes history records for a deleted DNS policy.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPDNSPolicy.Update().
		Where(
			bronzehistorygcpdnspolicy.ResourceID(resourceID),
			bronzehistorygcpdnspolicy.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if ent.IsNotFound(err) {
		return nil // No history to close
	}
	return err
}
