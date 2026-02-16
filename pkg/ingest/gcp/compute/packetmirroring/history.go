package packetmirroring

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcomputepacketmirroring"
)

// HistoryService manages packet mirroring history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new packet mirroring.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *PacketMirroringData, now time.Time) error {
	create := tx.BronzeHistoryGCPComputePacketMirroring.Create().
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
	if data.Region != "" {
		create.SetRegion(data.Region)
	}
	if data.Network != "" {
		create.SetNetwork(data.Network)
	}
	if data.Priority != 0 {
		create.SetPriority(data.Priority)
	}
	if data.Enable != "" {
		create.SetEnable(data.Enable)
	}
	if data.CollectorIlbJSON != nil {
		create.SetCollectorIlbJSON(data.CollectorIlbJSON)
	}
	if data.MirroredResourcesJSON != nil {
		create.SetMirroredResourcesJSON(data.MirroredResourcesJSON)
	}
	if data.FilterJSON != nil {
		create.SetFilterJSON(data.FilterJSON)
	}
	if data.CreationTimestamp != "" {
		create.SetCreationTimestamp(data.CreationTimestamp)
	}

	_, err := create.Save(ctx)
	return err
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputePacketMirroring, new *PacketMirroringData, diff *PacketMirroringDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	_, err := tx.BronzeHistoryGCPComputePacketMirroring.Update().
		Where(
			bronzehistorygcpcomputepacketmirroring.ResourceID(old.ID),
			bronzehistorygcpcomputepacketmirroring.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close old history: %w", err)
	}

	// Create new history
	create := tx.BronzeHistoryGCPComputePacketMirroring.Create().
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
	if new.Region != "" {
		create.SetRegion(new.Region)
	}
	if new.Network != "" {
		create.SetNetwork(new.Network)
	}
	if new.Priority != 0 {
		create.SetPriority(new.Priority)
	}
	if new.Enable != "" {
		create.SetEnable(new.Enable)
	}
	if new.CollectorIlbJSON != nil {
		create.SetCollectorIlbJSON(new.CollectorIlbJSON)
	}
	if new.MirroredResourcesJSON != nil {
		create.SetMirroredResourcesJSON(new.MirroredResourcesJSON)
	}
	if new.FilterJSON != nil {
		create.SetFilterJSON(new.FilterJSON)
	}
	if new.CreationTimestamp != "" {
		create.SetCreationTimestamp(new.CreationTimestamp)
	}

	_, err = create.Save(ctx)
	return err
}

// CloseHistory closes history records for a deleted packet mirroring.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPComputePacketMirroring.Update().
		Where(
			bronzehistorygcpcomputepacketmirroring.ResourceID(resourceID),
			bronzehistorygcpcomputepacketmirroring.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if ent.IsNotFound(err) {
		return nil // No history to close
	}
	return err
}
