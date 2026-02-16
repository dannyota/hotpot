package router

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcomputerouter"
)

// HistoryService manages router history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new router.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *RouterData, now time.Time) error {
	create := tx.BronzeHistoryGCPComputeRouter.Create().
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
	if data.Network != "" {
		create.SetNetwork(data.Network)
	}
	if data.Region != "" {
		create.SetRegion(data.Region)
	}
	if data.BgpAsn != 0 {
		create.SetBgpAsn(data.BgpAsn)
	}
	if data.BgpAdvertiseMode != "" {
		create.SetBgpAdvertiseMode(data.BgpAdvertiseMode)
	}
	if data.BgpAdvertisedGroupsJSON != nil {
		create.SetBgpAdvertisedGroupsJSON(data.BgpAdvertisedGroupsJSON)
	}
	if data.BgpAdvertisedIPRangesJSON != nil {
		create.SetBgpAdvertisedIPRangesJSON(data.BgpAdvertisedIPRangesJSON)
	}
	if data.BgpKeepaliveInterval != 0 {
		create.SetBgpKeepaliveInterval(data.BgpKeepaliveInterval)
	}
	if data.BgpPeersJSON != nil {
		create.SetBgpPeersJSON(data.BgpPeersJSON)
	}
	if data.InterfacesJSON != nil {
		create.SetInterfacesJSON(data.InterfacesJSON)
	}
	if data.NatsJSON != nil {
		create.SetNatsJSON(data.NatsJSON)
	}
	if data.EncryptedInterconnectRouter {
		create.SetEncryptedInterconnectRouter(data.EncryptedInterconnectRouter)
	}

	_, err := create.Save(ctx)
	return err
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeRouter, new *RouterData, diff *RouterDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	_, err := tx.BronzeHistoryGCPComputeRouter.Update().
		Where(
			bronzehistorygcpcomputerouter.ResourceID(old.ID),
			bronzehistorygcpcomputerouter.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close old history: %w", err)
	}

	// Create new history
	create := tx.BronzeHistoryGCPComputeRouter.Create().
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
	if new.Network != "" {
		create.SetNetwork(new.Network)
	}
	if new.Region != "" {
		create.SetRegion(new.Region)
	}
	if new.BgpAsn != 0 {
		create.SetBgpAsn(new.BgpAsn)
	}
	if new.BgpAdvertiseMode != "" {
		create.SetBgpAdvertiseMode(new.BgpAdvertiseMode)
	}
	if new.BgpAdvertisedGroupsJSON != nil {
		create.SetBgpAdvertisedGroupsJSON(new.BgpAdvertisedGroupsJSON)
	}
	if new.BgpAdvertisedIPRangesJSON != nil {
		create.SetBgpAdvertisedIPRangesJSON(new.BgpAdvertisedIPRangesJSON)
	}
	if new.BgpKeepaliveInterval != 0 {
		create.SetBgpKeepaliveInterval(new.BgpKeepaliveInterval)
	}
	if new.BgpPeersJSON != nil {
		create.SetBgpPeersJSON(new.BgpPeersJSON)
	}
	if new.InterfacesJSON != nil {
		create.SetInterfacesJSON(new.InterfacesJSON)
	}
	if new.NatsJSON != nil {
		create.SetNatsJSON(new.NatsJSON)
	}
	if new.EncryptedInterconnectRouter {
		create.SetEncryptedInterconnectRouter(new.EncryptedInterconnectRouter)
	}

	_, err = create.Save(ctx)
	return err
}

// CloseHistory closes history records for a deleted router.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPComputeRouter.Update().
		Where(
			bronzehistorygcpcomputerouter.ResourceID(resourceID),
			bronzehistorygcpcomputerouter.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if ent.IsNotFound(err) {
		return nil // No history to close
	}
	return err
}
