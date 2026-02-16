package firewall

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcomputefirewall"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcomputefirewallallowed"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcomputefirewalldenied"
)

// HistoryService handles history tracking for firewalls.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new firewall and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, firewallData *FirewallData, now time.Time) error {
	// Create firewall history
	fwHist, err := tx.BronzeHistoryGCPComputeFirewall.Create().
		SetResourceID(firewallData.ID).
		SetValidFrom(now).
		SetCollectedAt(firewallData.CollectedAt).
		SetFirstCollectedAt(firewallData.CollectedAt).
		SetName(firewallData.Name).
		SetDescription(firewallData.Description).
		SetSelfLink(firewallData.SelfLink).
		SetCreationTimestamp(firewallData.CreationTimestamp).
		SetNetwork(firewallData.Network).
		SetPriority(firewallData.Priority).
		SetDirection(firewallData.Direction).
		SetDisabled(firewallData.Disabled).
		SetSourceRangesJSON(firewallData.SourceRangesJSON).
		SetDestinationRangesJSON(firewallData.DestinationRangesJSON).
		SetSourceTagsJSON(firewallData.SourceTagsJSON).
		SetTargetTagsJSON(firewallData.TargetTagsJSON).
		SetSourceServiceAccountsJSON(firewallData.SourceServiceAccountsJSON).
		SetTargetServiceAccountsJSON(firewallData.TargetServiceAccountsJSON).
		SetLogConfigJSON(firewallData.LogConfigJSON).
		SetProjectID(firewallData.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create firewall history: %w", err)
	}

	// Create children history with firewall_history_id
	return h.createChildrenHistory(ctx, tx, fwHist.HistoryID, firewallData, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeFirewall, new *FirewallData, diff *FirewallDiff, now time.Time) error {
	// Get current firewall history
	currentHist, err := tx.BronzeHistoryGCPComputeFirewall.Query().
		Where(
			bronzehistorygcpcomputefirewall.ResourceID(old.ID),
			bronzehistorygcpcomputefirewall.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current firewall history: %w", err)
	}

	// If firewall-level fields changed, close old and create new firewall history
	if diff.IsChanged {
		// Close old firewall history
		err = tx.BronzeHistoryGCPComputeFirewall.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current firewall history: %w", err)
		}

		// Create new firewall history
		fwHist, err := tx.BronzeHistoryGCPComputeFirewall.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetDescription(new.Description).
			SetSelfLink(new.SelfLink).
			SetCreationTimestamp(new.CreationTimestamp).
			SetNetwork(new.Network).
			SetPriority(new.Priority).
			SetDirection(new.Direction).
			SetDisabled(new.Disabled).
			SetSourceRangesJSON(new.SourceRangesJSON).
			SetDestinationRangesJSON(new.DestinationRangesJSON).
			SetSourceTagsJSON(new.SourceTagsJSON).
			SetTargetTagsJSON(new.TargetTagsJSON).
			SetSourceServiceAccountsJSON(new.SourceServiceAccountsJSON).
			SetTargetServiceAccountsJSON(new.TargetServiceAccountsJSON).
			SetLogConfigJSON(new.LogConfigJSON).
			SetProjectID(new.ProjectID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new firewall history: %w", err)
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(ctx, tx, currentHist.HistoryID, now); err != nil {
			return fmt.Errorf("failed to close children history: %w", err)
		}
		return h.createChildrenHistory(ctx, tx, fwHist.HistoryID, new, now)
	}

	// Firewall unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(ctx, tx, currentHist.HistoryID, new, diff, now)
}

// CloseHistory closes history records for a deleted firewall.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	// Get current firewall history
	currentHist, err := tx.BronzeHistoryGCPComputeFirewall.Query().
		Where(
			bronzehistorygcpcomputefirewall.ResourceID(resourceID),
			bronzehistorygcpcomputefirewall.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("failed to find current firewall history: %w", err)
	}

	// Close firewall history
	err = tx.BronzeHistoryGCPComputeFirewall.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close firewall history: %w", err)
	}

	// Close all children history
	return h.closeChildrenHistory(ctx, tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(ctx context.Context, tx *ent.Tx, firewallHistoryID uint, firewall *FirewallData, now time.Time) error {
	// Allowed rules
	for _, allowed := range firewall.Allowed {
		create := tx.BronzeHistoryGCPComputeFirewallAllowed.Create().
			SetFirewallHistoryID(firewallHistoryID).
			SetValidFrom(now).
			SetIPProtocol(allowed.IpProtocol)
		if allowed.PortsJSON != nil {
			create.SetPortsJSON(allowed.PortsJSON)
		}
		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create allowed history: %w", err)
		}
	}

	// Denied rules
	for _, denied := range firewall.Denied {
		create := tx.BronzeHistoryGCPComputeFirewallDenied.Create().
			SetFirewallHistoryID(firewallHistoryID).
			SetValidFrom(now).
			SetIPProtocol(denied.IpProtocol)
		if denied.PortsJSON != nil {
			create.SetPortsJSON(denied.PortsJSON)
		}
		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create denied history: %w", err)
		}
	}

	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(ctx context.Context, tx *ent.Tx, firewallHistoryID uint, now time.Time) error {
	// Close allowed rules
	_, err := tx.BronzeHistoryGCPComputeFirewallAllowed.Update().
		Where(
			bronzehistorygcpcomputefirewallallowed.FirewallHistoryID(firewallHistoryID),
			bronzehistorygcpcomputefirewallallowed.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close allowed history: %w", err)
	}

	// Close denied rules
	_, err = tx.BronzeHistoryGCPComputeFirewallDenied.Update().
		Where(
			bronzehistorygcpcomputefirewalldenied.FirewallHistoryID(firewallHistoryID),
			bronzehistorygcpcomputefirewalldenied.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close denied history: %w", err)
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(ctx context.Context, tx *ent.Tx, firewallHistoryID uint, new *FirewallData, diff *FirewallDiff, now time.Time) error {
	if diff.AllowedDiff.Changed {
		if err := h.updateAllowedHistory(ctx, tx, firewallHistoryID, new.Allowed, now); err != nil {
			return fmt.Errorf("failed to update allowed history: %w", err)
		}
	}

	if diff.DeniedDiff.Changed {
		if err := h.updateDeniedHistory(ctx, tx, firewallHistoryID, new.Denied, now); err != nil {
			return fmt.Errorf("failed to update denied history: %w", err)
		}
	}

	return nil
}

func (h *HistoryService) updateAllowedHistory(ctx context.Context, tx *ent.Tx, firewallHistoryID uint, allowed []AllowedData, now time.Time) error {
	// Close old allowed history
	_, err := tx.BronzeHistoryGCPComputeFirewallAllowed.Update().
		Where(
			bronzehistorygcpcomputefirewallallowed.FirewallHistoryID(firewallHistoryID),
			bronzehistorygcpcomputefirewallallowed.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close allowed history: %w", err)
	}

	// Create new allowed history
	for _, a := range allowed {
		create := tx.BronzeHistoryGCPComputeFirewallAllowed.Create().
			SetFirewallHistoryID(firewallHistoryID).
			SetValidFrom(now).
			SetIPProtocol(a.IpProtocol)
		if a.PortsJSON != nil {
			create.SetPortsJSON(a.PortsJSON)
		}
		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create allowed history: %w", err)
		}
	}

	return nil
}

func (h *HistoryService) updateDeniedHistory(ctx context.Context, tx *ent.Tx, firewallHistoryID uint, denied []DeniedData, now time.Time) error {
	// Close old denied history
	_, err := tx.BronzeHistoryGCPComputeFirewallDenied.Update().
		Where(
			bronzehistorygcpcomputefirewalldenied.FirewallHistoryID(firewallHistoryID),
			bronzehistorygcpcomputefirewalldenied.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close denied history: %w", err)
	}

	// Create new denied history
	for _, d := range denied {
		create := tx.BronzeHistoryGCPComputeFirewallDenied.Create().
			SetFirewallHistoryID(firewallHistoryID).
			SetValidFrom(now).
			SetIPProtocol(d.IpProtocol)
		if d.PortsJSON != nil {
			create.SetPortsJSON(d.PortsJSON)
		}
		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create denied history: %w", err)
		}
	}

	return nil
}
