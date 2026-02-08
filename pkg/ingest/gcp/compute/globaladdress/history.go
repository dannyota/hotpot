package globaladdress

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputeglobaladdress"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputeglobaladdresslabel"
)

// HistoryService handles history tracking for global addresses.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new global address and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, addressData *GlobalAddressData, now time.Time) error {
	// Create address history
	addrHistCreate := tx.BronzeHistoryGCPComputeGlobalAddress.Create().
		SetResourceID(addressData.ID).
		SetValidFrom(now).
		SetCollectedAt(addressData.CollectedAt).
		SetName(addressData.Name).
		SetDescription(addressData.Description).
		SetAddress(addressData.Address).
		SetAddressType(addressData.AddressType).
		SetIPVersion(addressData.IpVersion).
		SetIpv6EndpointType(addressData.Ipv6EndpointType).
		SetIPCollection(addressData.IpCollection).
		SetRegion(addressData.Region).
		SetStatus(addressData.Status).
		SetPurpose(addressData.Purpose).
		SetNetwork(addressData.Network).
		SetSubnetwork(addressData.Subnetwork).
		SetNetworkTier(addressData.NetworkTier).
		SetPrefixLength(addressData.PrefixLength).
		SetSelfLink(addressData.SelfLink).
		SetCreationTimestamp(addressData.CreationTimestamp).
		SetLabelFingerprint(addressData.LabelFingerprint).
		SetProjectID(addressData.ProjectID)

	if addressData.UsersJSON != nil {
		addrHistCreate.SetUsersJSON(addressData.UsersJSON)
	}

	addrHist, err := addrHistCreate.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create global address history: %w", err)
	}

	// Create children history with global_address_history_id
	return h.createChildrenHistory(ctx, tx, addrHist.HistoryID, addressData, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeGlobalAddress, new *GlobalAddressData, diff *GlobalAddressDiff, now time.Time) error {
	// Get current address history
	currentHist, err := tx.BronzeHistoryGCPComputeGlobalAddress.Query().
		Where(
			bronzehistorygcpcomputeglobaladdress.ResourceID(old.ID),
			bronzehistorygcpcomputeglobaladdress.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current global address history: %w", err)
	}

	// If address-level fields changed, close old and create new address history
	if diff.IsChanged {
		// Close old address history
		if err := tx.BronzeHistoryGCPComputeGlobalAddress.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to close global address history: %w", err)
		}

		// Create new address history
		addrHistCreate := tx.BronzeHistoryGCPComputeGlobalAddress.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetName(new.Name).
			SetDescription(new.Description).
			SetAddress(new.Address).
			SetAddressType(new.AddressType).
			SetIPVersion(new.IpVersion).
			SetIpv6EndpointType(new.Ipv6EndpointType).
			SetIPCollection(new.IpCollection).
			SetRegion(new.Region).
			SetStatus(new.Status).
			SetPurpose(new.Purpose).
			SetNetwork(new.Network).
			SetSubnetwork(new.Subnetwork).
			SetNetworkTier(new.NetworkTier).
			SetPrefixLength(new.PrefixLength).
			SetSelfLink(new.SelfLink).
			SetCreationTimestamp(new.CreationTimestamp).
			SetLabelFingerprint(new.LabelFingerprint).
			SetProjectID(new.ProjectID)

		if new.UsersJSON != nil {
			addrHistCreate.SetUsersJSON(new.UsersJSON)
		}

		addrHist, err := addrHistCreate.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new global address history: %w", err)
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(ctx, tx, currentHist.HistoryID, now); err != nil {
			return fmt.Errorf("failed to close children history: %w", err)
		}
		return h.createChildrenHistory(ctx, tx, addrHist.HistoryID, new, now)
	}

	// Address unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(ctx, tx, currentHist.HistoryID, new, diff, now)
}

// CloseHistory closes history records for a deleted global address.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	// Get current address history
	currentHist, err := tx.BronzeHistoryGCPComputeGlobalAddress.Query().
		Where(
			bronzehistorygcpcomputeglobaladdress.ResourceID(resourceID),
			bronzehistorygcpcomputeglobaladdress.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("failed to find current global address history: %w", err)
	}

	// Close address history
	if err := tx.BronzeHistoryGCPComputeGlobalAddress.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("failed to close global address history: %w", err)
	}

	// Close all children history
	return h.closeChildrenHistory(ctx, tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(ctx context.Context, tx *ent.Tx, globalAddressHistoryID uint, data *GlobalAddressData, now time.Time) error {
	// Labels
	for _, labelData := range data.Labels {
		_, err := tx.BronzeHistoryGCPComputeGlobalAddressLabel.Create().
			SetGlobalAddressHistoryID(globalAddressHistoryID).
			SetValidFrom(now).
			SetKey(labelData.Key).
			SetValue(labelData.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label history: %w", err)
		}
	}

	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(ctx context.Context, tx *ent.Tx, globalAddressHistoryID uint, now time.Time) error {
	// Close labels
	_, err := tx.BronzeHistoryGCPComputeGlobalAddressLabel.Update().
		Where(
			bronzehistorygcpcomputeglobaladdresslabel.GlobalAddressHistoryID(globalAddressHistoryID),
			bronzehistorygcpcomputeglobaladdresslabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close labels history: %w", err)
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(ctx context.Context, tx *ent.Tx, globalAddressHistoryID uint, new *GlobalAddressData, diff *GlobalAddressDiff, now time.Time) error {
	if diff.LabelsDiff.Changed {
		if err := h.updateLabelsHistory(ctx, tx, globalAddressHistoryID, new.Labels, now); err != nil {
			return err
		}
	}
	return nil
}

func (h *HistoryService) updateLabelsHistory(ctx context.Context, tx *ent.Tx, globalAddressHistoryID uint, labels []GlobalAddressLabelData, now time.Time) error {
	// Close old labels
	_, err := tx.BronzeHistoryGCPComputeGlobalAddressLabel.Update().
		Where(
			bronzehistorygcpcomputeglobaladdresslabel.GlobalAddressHistoryID(globalAddressHistoryID),
			bronzehistorygcpcomputeglobaladdresslabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close labels history: %w", err)
	}

	// Create new labels
	for _, labelData := range labels {
		_, err := tx.BronzeHistoryGCPComputeGlobalAddressLabel.Create().
			SetGlobalAddressHistoryID(globalAddressHistoryID).
			SetValidFrom(now).
			SetKey(labelData.Key).
			SetValue(labelData.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label history: %w", err)
		}
	}

	return nil
}
