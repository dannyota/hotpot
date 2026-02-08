package instancegroup

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputeinstancegroup"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputeinstancegroupmember"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputeinstancegroupnamedport"
)

// HistoryService manages instance group history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new instance group.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, groupData *InstanceGroupData, now time.Time) error {
	// Create instance group history
	groupHistory, err := tx.BronzeHistoryGCPComputeInstanceGroup.Create().
		SetResourceID(groupData.ID).
		SetValidFrom(now).
		SetCollectedAt(groupData.CollectedAt).
		SetFirstCollectedAt(groupData.CollectedAt).
		SetName(groupData.Name).
		SetDescription(groupData.Description).
		SetZone(groupData.Zone).
		SetNetwork(groupData.Network).
		SetSubnetwork(groupData.Subnetwork).
		SetSize(groupData.Size).
		SetSelfLink(groupData.SelfLink).
		SetCreationTimestamp(groupData.CreationTimestamp).
		SetFingerprint(groupData.Fingerprint).
		SetProjectID(groupData.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create instance group history: %w", err)
	}

	// Create named port history
	for _, port := range groupData.NamedPorts {
		_, err := tx.BronzeHistoryGCPComputeInstanceGroupNamedPort.Create().
			SetGroupHistoryID(groupHistory.HistoryID).
			SetValidFrom(now).
			SetName(port.Name).
			SetPort(port.Port).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create named port history: %w", err)
		}
	}

	// Create member history
	for _, member := range groupData.Members {
		_, err := tx.BronzeHistoryGCPComputeInstanceGroupMember.Create().
			SetGroupHistoryID(groupHistory.HistoryID).
			SetValidFrom(now).
			SetInstanceURL(member.InstanceURL).
			SetInstanceName(member.InstanceName).
			SetStatus(member.Status).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create member history: %w", err)
		}
	}

	return nil
}

// UpdateHistory updates history records for a changed instance group.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeInstanceGroup, new *InstanceGroupData, diff *InstanceGroupDiff, now time.Time) error {
	// Get current instance group history
	currentHistory, err := tx.BronzeHistoryGCPComputeInstanceGroup.Query().
		Where(
			bronzehistorygcpcomputeinstancegroup.ResourceID(old.ID),
			bronzehistorygcpcomputeinstancegroup.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current instance group history: %w", err)
	}

	// Close current instance group history if core fields changed
	if diff.IsChanged {
		// Close old children history first
		if err := h.closeChildrenHistory(ctx, tx, currentHistory.HistoryID, now); err != nil {
			return err
		}

		// Close current instance group history
		err = tx.BronzeHistoryGCPComputeInstanceGroup.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current instance group history: %w", err)
		}

		// Create new instance group history
		newHistory, err := tx.BronzeHistoryGCPComputeInstanceGroup.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetDescription(new.Description).
			SetZone(new.Zone).
			SetNetwork(new.Network).
			SetSubnetwork(new.Subnetwork).
			SetSize(new.Size).
			SetSelfLink(new.SelfLink).
			SetCreationTimestamp(new.CreationTimestamp).
			SetFingerprint(new.Fingerprint).
			SetProjectID(new.ProjectID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new instance group history: %w", err)
		}

		// Create new children history linked to new group history
		return h.createChildrenHistory(ctx, tx, newHistory.HistoryID, new, now)
	}

	// Instance group unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(ctx, tx, currentHistory.HistoryID, new, diff, now)
}

// CloseHistory closes all history records for a deleted instance group.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	// Get current instance group history
	currentHistory, err := tx.BronzeHistoryGCPComputeInstanceGroup.Query().
		Where(
			bronzehistorygcpcomputeinstancegroup.ResourceID(resourceID),
			bronzehistorygcpcomputeinstancegroup.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("failed to find current instance group history: %w", err)
	}

	// Close instance group history
	err = tx.BronzeHistoryGCPComputeInstanceGroup.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close instance group history: %w", err)
	}

	// Close all children history
	return h.closeChildrenHistory(ctx, tx, currentHistory.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(ctx context.Context, tx *ent.Tx, groupHistoryID uint, groupData *InstanceGroupData, now time.Time) error {
	// Create named port history
	for _, port := range groupData.NamedPorts {
		_, err := tx.BronzeHistoryGCPComputeInstanceGroupNamedPort.Create().
			SetGroupHistoryID(groupHistoryID).
			SetValidFrom(now).
			SetName(port.Name).
			SetPort(port.Port).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create named port history: %w", err)
		}
	}

	// Create member history
	for _, member := range groupData.Members {
		_, err := tx.BronzeHistoryGCPComputeInstanceGroupMember.Create().
			SetGroupHistoryID(groupHistoryID).
			SetValidFrom(now).
			SetInstanceURL(member.InstanceURL).
			SetInstanceName(member.InstanceName).
			SetStatus(member.Status).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create member history: %w", err)
		}
	}

	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(ctx context.Context, tx *ent.Tx, groupHistoryID uint, now time.Time) error {
	// Close named port history
	_, err := tx.BronzeHistoryGCPComputeInstanceGroupNamedPort.Update().
		Where(
			bronzehistorygcpcomputeinstancegroupnamedport.GroupHistoryID(groupHistoryID),
			bronzehistorygcpcomputeinstancegroupnamedport.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close named port history: %w", err)
	}

	// Close member history
	_, err = tx.BronzeHistoryGCPComputeInstanceGroupMember.Update().
		Where(
			bronzehistorygcpcomputeinstancegroupmember.GroupHistoryID(groupHistoryID),
			bronzehistorygcpcomputeinstancegroupmember.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close member history: %w", err)
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(ctx context.Context, tx *ent.Tx, groupHistoryID uint, new *InstanceGroupData, diff *InstanceGroupDiff, now time.Time) error {
	if diff.NamedPortsDiff.HasChanges {
		// Close old named port history
		_, err := tx.BronzeHistoryGCPComputeInstanceGroupNamedPort.Update().
			Where(
				bronzehistorygcpcomputeinstancegroupnamedport.GroupHistoryID(groupHistoryID),
				bronzehistorygcpcomputeinstancegroupnamedport.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to close named port history: %w", err)
		}

		// Create new named port history
		for _, port := range new.NamedPorts {
			_, err := tx.BronzeHistoryGCPComputeInstanceGroupNamedPort.Create().
				SetGroupHistoryID(groupHistoryID).
				SetValidFrom(now).
				SetName(port.Name).
				SetPort(port.Port).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create named port history: %w", err)
			}
		}
	}

	if diff.MembersDiff.HasChanges {
		// Close old member history
		_, err := tx.BronzeHistoryGCPComputeInstanceGroupMember.Update().
			Where(
				bronzehistorygcpcomputeinstancegroupmember.GroupHistoryID(groupHistoryID),
				bronzehistorygcpcomputeinstancegroupmember.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to close member history: %w", err)
		}

		// Create new member history
		for _, member := range new.Members {
			_, err := tx.BronzeHistoryGCPComputeInstanceGroupMember.Create().
				SetGroupHistoryID(groupHistoryID).
				SetValidFrom(now).
				SetInstanceURL(member.InstanceURL).
				SetInstanceName(member.InstanceName).
				SetStatus(member.Status).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create member history: %w", err)
			}
		}
	}

	return nil
}
