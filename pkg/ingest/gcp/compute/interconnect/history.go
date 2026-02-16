package interconnect

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcomputeinterconnect"
)

// HistoryService manages interconnect history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new interconnect.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *InterconnectData, now time.Time) error {
	create := tx.BronzeHistoryGCPComputeInterconnect.Create().
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
	if data.Location != "" {
		create.SetLocation(data.Location)
	}
	if data.InterconnectType != "" {
		create.SetInterconnectType(data.InterconnectType)
	}
	if data.LinkType != "" {
		create.SetLinkType(data.LinkType)
	}
	if data.AdminEnabled {
		create.SetAdminEnabled(data.AdminEnabled)
	}
	if data.OperationalStatus != "" {
		create.SetOperationalStatus(data.OperationalStatus)
	}
	if data.ProvisionedLinkCount != 0 {
		create.SetProvisionedLinkCount(data.ProvisionedLinkCount)
	}
	if data.RequestedLinkCount != 0 {
		create.SetRequestedLinkCount(data.RequestedLinkCount)
	}
	if data.PeerIPAddress != "" {
		create.SetPeerIPAddress(data.PeerIPAddress)
	}
	if data.GoogleIPAddress != "" {
		create.SetGoogleIPAddress(data.GoogleIPAddress)
	}
	if data.GoogleReferenceID != "" {
		create.SetGoogleReferenceID(data.GoogleReferenceID)
	}
	if data.NocContactEmail != "" {
		create.SetNocContactEmail(data.NocContactEmail)
	}
	if data.CustomerName != "" {
		create.SetCustomerName(data.CustomerName)
	}
	if data.State != "" {
		create.SetState(data.State)
	}
	if data.CreationTimestamp != "" {
		create.SetCreationTimestamp(data.CreationTimestamp)
	}
	if data.ExpectedOutagesJSON != nil {
		create.SetExpectedOutagesJSON(data.ExpectedOutagesJSON)
	}
	if data.CircuitInfosJSON != nil {
		create.SetCircuitInfosJSON(data.CircuitInfosJSON)
	}

	_, err := create.Save(ctx)
	return err
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeInterconnect, new *InterconnectData, diff *InterconnectDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	_, err := tx.BronzeHistoryGCPComputeInterconnect.Update().
		Where(
			bronzehistorygcpcomputeinterconnect.ResourceID(old.ID),
			bronzehistorygcpcomputeinterconnect.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close old history: %w", err)
	}

	// Create new history
	create := tx.BronzeHistoryGCPComputeInterconnect.Create().
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
	if new.Location != "" {
		create.SetLocation(new.Location)
	}
	if new.InterconnectType != "" {
		create.SetInterconnectType(new.InterconnectType)
	}
	if new.LinkType != "" {
		create.SetLinkType(new.LinkType)
	}
	if new.AdminEnabled {
		create.SetAdminEnabled(new.AdminEnabled)
	}
	if new.OperationalStatus != "" {
		create.SetOperationalStatus(new.OperationalStatus)
	}
	if new.ProvisionedLinkCount != 0 {
		create.SetProvisionedLinkCount(new.ProvisionedLinkCount)
	}
	if new.RequestedLinkCount != 0 {
		create.SetRequestedLinkCount(new.RequestedLinkCount)
	}
	if new.PeerIPAddress != "" {
		create.SetPeerIPAddress(new.PeerIPAddress)
	}
	if new.GoogleIPAddress != "" {
		create.SetGoogleIPAddress(new.GoogleIPAddress)
	}
	if new.GoogleReferenceID != "" {
		create.SetGoogleReferenceID(new.GoogleReferenceID)
	}
	if new.NocContactEmail != "" {
		create.SetNocContactEmail(new.NocContactEmail)
	}
	if new.CustomerName != "" {
		create.SetCustomerName(new.CustomerName)
	}
	if new.State != "" {
		create.SetState(new.State)
	}
	if new.CreationTimestamp != "" {
		create.SetCreationTimestamp(new.CreationTimestamp)
	}
	if new.ExpectedOutagesJSON != nil {
		create.SetExpectedOutagesJSON(new.ExpectedOutagesJSON)
	}
	if new.CircuitInfosJSON != nil {
		create.SetCircuitInfosJSON(new.CircuitInfosJSON)
	}

	_, err = create.Save(ctx)
	return err
}

// CloseHistory closes history records for a deleted interconnect.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPComputeInterconnect.Update().
		Where(
			bronzehistorygcpcomputeinterconnect.ResourceID(resourceID),
			bronzehistorygcpcomputeinterconnect.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if ent.IsNotFound(err) {
		return nil // No history to close
	}
	return err
}
