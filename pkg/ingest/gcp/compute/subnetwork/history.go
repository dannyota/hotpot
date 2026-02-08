package subnetwork

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputesubnetwork"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputesubnetworksecondaryrange"
)

// HistoryService handles history tracking for subnetworks.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new subnetwork and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, subnetData *SubnetworkData, now time.Time) error {
	// Create subnetwork history
	subnetHist, err := tx.BronzeHistoryGCPComputeSubnetwork.Create().
		SetResourceID(subnetData.ID).
		SetValidFrom(now).
		SetCollectedAt(subnetData.CollectedAt).
		SetName(subnetData.Name).
		SetDescription(subnetData.Description).
		SetSelfLink(subnetData.SelfLink).
		SetCreationTimestamp(subnetData.CreationTimestamp).
		SetNetwork(subnetData.Network).
		SetRegion(subnetData.Region).
		SetIPCidrRange(subnetData.IpCidrRange).
		SetGatewayAddress(subnetData.GatewayAddress).
		SetPurpose(subnetData.Purpose).
		SetRole(subnetData.Role).
		SetPrivateIPGoogleAccess(subnetData.PrivateIpGoogleAccess).
		SetPrivateIpv6GoogleAccess(subnetData.PrivateIpv6GoogleAccess).
		SetStackType(subnetData.StackType).
		SetIpv6AccessType(subnetData.Ipv6AccessType).
		SetInternalIpv6Prefix(subnetData.InternalIpv6Prefix).
		SetExternalIpv6Prefix(subnetData.ExternalIpv6Prefix).
		SetLogConfigJSON(subnetData.LogConfigJSON).
		SetFingerprint(subnetData.Fingerprint).
		SetProjectID(subnetData.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create subnetwork history: %w", err)
	}

	// Create children history with subnetwork_history_id
	return h.createChildrenHistory(ctx, tx, subnetHist.HistoryID, subnetData, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeSubnetwork, new *SubnetworkData, diff *SubnetworkDiff, now time.Time) error {
	// Get current subnetwork history
	currentHist, err := tx.BronzeHistoryGCPComputeSubnetwork.Query().
		Where(
			bronzehistorygcpcomputesubnetwork.ResourceID(old.ID),
			bronzehistorygcpcomputesubnetwork.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current subnetwork history: %w", err)
	}

	// If subnetwork-level fields changed, close old and create new subnetwork history
	if diff.IsChanged {
		// Close old subnetwork history
		err = tx.BronzeHistoryGCPComputeSubnetwork.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current subnetwork history: %w", err)
		}

		// Create new subnetwork history
		subnetHist, err := tx.BronzeHistoryGCPComputeSubnetwork.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetName(new.Name).
			SetDescription(new.Description).
			SetSelfLink(new.SelfLink).
			SetCreationTimestamp(new.CreationTimestamp).
			SetNetwork(new.Network).
			SetRegion(new.Region).
			SetIPCidrRange(new.IpCidrRange).
			SetGatewayAddress(new.GatewayAddress).
			SetPurpose(new.Purpose).
			SetRole(new.Role).
			SetPrivateIPGoogleAccess(new.PrivateIpGoogleAccess).
			SetPrivateIpv6GoogleAccess(new.PrivateIpv6GoogleAccess).
			SetStackType(new.StackType).
			SetIpv6AccessType(new.Ipv6AccessType).
			SetInternalIpv6Prefix(new.InternalIpv6Prefix).
			SetExternalIpv6Prefix(new.ExternalIpv6Prefix).
			SetLogConfigJSON(new.LogConfigJSON).
			SetFingerprint(new.Fingerprint).
			SetProjectID(new.ProjectID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new subnetwork history: %w", err)
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(ctx, tx, currentHist.HistoryID, now); err != nil {
			return fmt.Errorf("failed to close children history: %w", err)
		}
		return h.createChildrenHistory(ctx, tx, subnetHist.HistoryID, new, now)
	}

	// Subnetwork unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(ctx, tx, currentHist.HistoryID, new, diff, now)
}

// CloseHistory closes history records for a deleted subnetwork.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	// Get current subnetwork history
	currentHist, err := tx.BronzeHistoryGCPComputeSubnetwork.Query().
		Where(
			bronzehistorygcpcomputesubnetwork.ResourceID(resourceID),
			bronzehistorygcpcomputesubnetwork.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("failed to find current subnetwork history: %w", err)
	}

	// Close subnetwork history
	err = tx.BronzeHistoryGCPComputeSubnetwork.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close subnetwork history: %w", err)
	}

	// Close all children history
	return h.closeChildrenHistory(ctx, tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(ctx context.Context, tx *ent.Tx, subnetHistoryID uint, subnet *SubnetworkData, now time.Time) error {
	// Secondary ranges
	for _, rangeData := range subnet.SecondaryIpRanges {
		_, err := tx.BronzeHistoryGCPComputeSubnetworkSecondaryRange.Create().
			SetSubnetworkHistoryID(subnetHistoryID).
			SetValidFrom(now).
			SetRangeName(rangeData.RangeName).
			SetIPCidrRange(rangeData.IpCidrRange).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create secondary range history: %w", err)
		}
	}

	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(ctx context.Context, tx *ent.Tx, subnetHistoryID uint, now time.Time) error {
	// Close secondary ranges
	_, err := tx.BronzeHistoryGCPComputeSubnetworkSecondaryRange.Update().
		Where(
			bronzehistorygcpcomputesubnetworksecondaryrange.SubnetworkHistoryID(subnetHistoryID),
			bronzehistorygcpcomputesubnetworksecondaryrange.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close secondary range history: %w", err)
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(ctx context.Context, tx *ent.Tx, subnetHistoryID uint, new *SubnetworkData, diff *SubnetworkDiff, now time.Time) error {
	if diff.SecondaryRangesDiff.Changed {
		if err := h.updateSecondaryRangesHistory(ctx, tx, subnetHistoryID, new.SecondaryIpRanges, now); err != nil {
			return fmt.Errorf("failed to update secondary ranges history: %w", err)
		}
	}

	return nil
}

func (h *HistoryService) updateSecondaryRangesHistory(ctx context.Context, tx *ent.Tx, subnetHistoryID uint, ranges []SecondaryRangeData, now time.Time) error {
	// Close old secondary range history
	_, err := tx.BronzeHistoryGCPComputeSubnetworkSecondaryRange.Update().
		Where(
			bronzehistorygcpcomputesubnetworksecondaryrange.SubnetworkHistoryID(subnetHistoryID),
			bronzehistorygcpcomputesubnetworksecondaryrange.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close secondary range history: %w", err)
	}

	// Create new secondary range history
	for _, rangeData := range ranges {
		_, err := tx.BronzeHistoryGCPComputeSubnetworkSecondaryRange.Create().
			SetSubnetworkHistoryID(subnetHistoryID).
			SetValidFrom(now).
			SetRangeName(rangeData.RangeName).
			SetIPCidrRange(rangeData.IpCidrRange).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create secondary range history: %w", err)
		}
	}

	return nil
}
