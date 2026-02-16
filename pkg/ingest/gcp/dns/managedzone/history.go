package managedzone

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpdnsmanagedzone"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpdnsmanagedzonelabel"
)

// HistoryService handles history tracking for managed zones.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new managed zone and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, zoneData *ManagedZoneData, now time.Time) error {
	// Create managed zone history
	create := tx.BronzeHistoryGCPDNSManagedZone.Create().
		SetResourceID(zoneData.ID).
		SetValidFrom(now).
		SetCollectedAt(zoneData.CollectedAt).
		SetFirstCollectedAt(zoneData.CollectedAt).
		SetName(zoneData.Name).
		SetProjectID(zoneData.ProjectID)

	if zoneData.DnsName != "" {
		create.SetDNSName(zoneData.DnsName)
	}
	if zoneData.Description != "" {
		create.SetDescription(zoneData.Description)
	}
	if zoneData.Visibility != "" {
		create.SetVisibility(zoneData.Visibility)
	}
	if zoneData.CreationTime != "" {
		create.SetCreationTime(zoneData.CreationTime)
	}
	if zoneData.DnssecConfigJSON != nil {
		create.SetDnssecConfigJSON(zoneData.DnssecConfigJSON)
	}
	if zoneData.PrivateVisibilityConfigJSON != nil {
		create.SetPrivateVisibilityConfigJSON(zoneData.PrivateVisibilityConfigJSON)
	}
	if zoneData.ForwardingConfigJSON != nil {
		create.SetForwardingConfigJSON(zoneData.ForwardingConfigJSON)
	}
	if zoneData.PeeringConfigJSON != nil {
		create.SetPeeringConfigJSON(zoneData.PeeringConfigJSON)
	}
	if zoneData.CloudLoggingConfigJSON != nil {
		create.SetCloudLoggingConfigJSON(zoneData.CloudLoggingConfigJSON)
	}

	zoneHist, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create managed zone history: %w", err)
	}

	// Create children history with managed_zone_history_id
	return h.createLabelsHistory(ctx, tx, zoneHist.HistoryID, zoneData, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPDNSManagedZone, new *ManagedZoneData, diff *ManagedZoneDiff, now time.Time) error {
	// Get current managed zone history
	currentHist, err := tx.BronzeHistoryGCPDNSManagedZone.Query().
		Where(
			bronzehistorygcpdnsmanagedzone.ResourceID(old.ID),
			bronzehistorygcpdnsmanagedzone.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current managed zone history: %w", err)
	}

	// If managed zone-level fields changed, close old and create new history
	if diff.IsChanged {
		// Close old managed zone history
		err = tx.BronzeHistoryGCPDNSManagedZone.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current managed zone history: %w", err)
		}

		// Create new managed zone history
		create := tx.BronzeHistoryGCPDNSManagedZone.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetProjectID(new.ProjectID)

		if new.DnsName != "" {
			create.SetDNSName(new.DnsName)
		}
		if new.Description != "" {
			create.SetDescription(new.Description)
		}
		if new.Visibility != "" {
			create.SetVisibility(new.Visibility)
		}
		if new.CreationTime != "" {
			create.SetCreationTime(new.CreationTime)
		}
		if new.DnssecConfigJSON != nil {
			create.SetDnssecConfigJSON(new.DnssecConfigJSON)
		}
		if new.PrivateVisibilityConfigJSON != nil {
			create.SetPrivateVisibilityConfigJSON(new.PrivateVisibilityConfigJSON)
		}
		if new.ForwardingConfigJSON != nil {
			create.SetForwardingConfigJSON(new.ForwardingConfigJSON)
		}
		if new.PeeringConfigJSON != nil {
			create.SetPeeringConfigJSON(new.PeeringConfigJSON)
		}
		if new.CloudLoggingConfigJSON != nil {
			create.SetCloudLoggingConfigJSON(new.CloudLoggingConfigJSON)
		}

		zoneHist, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new managed zone history: %w", err)
		}

		// Close all children history and create new ones
		if err := h.closeLabelsHistory(ctx, tx, currentHist.HistoryID, now); err != nil {
			return fmt.Errorf("failed to close labels history: %w", err)
		}
		return h.createLabelsHistory(ctx, tx, zoneHist.HistoryID, new, now)
	}

	// Managed zone unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(ctx, tx, currentHist.HistoryID, new, diff, now)
}

// CloseHistory closes history records for a deleted managed zone.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	// Get current managed zone history
	currentHist, err := tx.BronzeHistoryGCPDNSManagedZone.Query().
		Where(
			bronzehistorygcpdnsmanagedzone.ResourceID(resourceID),
			bronzehistorygcpdnsmanagedzone.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("failed to find current managed zone history: %w", err)
	}

	// Close managed zone history
	err = tx.BronzeHistoryGCPDNSManagedZone.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close managed zone history: %w", err)
	}

	// Close all children history
	return h.closeLabelsHistory(ctx, tx, currentHist.HistoryID, now)
}

// createLabelsHistory creates history records for all labels.
func (h *HistoryService) createLabelsHistory(ctx context.Context, tx *ent.Tx, managedZoneHistoryID uint, zoneData *ManagedZoneData, now time.Time) error {
	for _, label := range zoneData.Labels {
		_, err := tx.BronzeHistoryGCPDNSManagedZoneLabel.Create().
			SetManagedZoneHistoryID(managedZoneHistoryID).
			SetValidFrom(now).
			SetKey(label.Key).
			SetValue(label.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label history: %w", err)
		}
	}

	return nil
}

// closeLabelsHistory closes all label history records.
func (h *HistoryService) closeLabelsHistory(ctx context.Context, tx *ent.Tx, managedZoneHistoryID uint, now time.Time) error {
	_, err := tx.BronzeHistoryGCPDNSManagedZoneLabel.Update().
		Where(
			bronzehistorygcpdnsmanagedzonelabel.ManagedZoneHistoryID(managedZoneHistoryID),
			bronzehistorygcpdnsmanagedzonelabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close label history: %w", err)
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(ctx context.Context, tx *ent.Tx, managedZoneHistoryID uint, new *ManagedZoneData, diff *ManagedZoneDiff, now time.Time) error {
	if diff.LabelDiff.Changed {
		if err := h.updateLabelsHistory(ctx, tx, managedZoneHistoryID, new.Labels, now); err != nil {
			return fmt.Errorf("failed to update labels history: %w", err)
		}
	}

	return nil
}

func (h *HistoryService) updateLabelsHistory(ctx context.Context, tx *ent.Tx, managedZoneHistoryID uint, labels []LabelData, now time.Time) error {
	// Close old label history
	_, err := tx.BronzeHistoryGCPDNSManagedZoneLabel.Update().
		Where(
			bronzehistorygcpdnsmanagedzonelabel.ManagedZoneHistoryID(managedZoneHistoryID),
			bronzehistorygcpdnsmanagedzonelabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close label history: %w", err)
	}

	// Create new label history
	for _, l := range labels {
		_, err := tx.BronzeHistoryGCPDNSManagedZoneLabel.Create().
			SetManagedZoneHistoryID(managedZoneHistoryID).
			SetValidFrom(now).
			SetKey(l.Key).
			SetValue(l.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label history: %w", err)
		}
	}

	return nil
}
