package disk

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputedisk"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputedisklabel"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputedisklicense"
)

// HistoryService handles history tracking for disks.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new disk and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, diskData *DiskData, now time.Time) error {
	// Create disk history
	diskHistCreate := tx.BronzeHistoryGCPComputeDisk.Create().
		SetResourceID(diskData.ID).
		SetValidFrom(now).
		SetCollectedAt(diskData.CollectedAt).
		SetFirstCollectedAt(diskData.CollectedAt).
		SetName(diskData.Name).
		SetDescription(diskData.Description).
		SetZone(diskData.Zone).
		SetRegion(diskData.Region).
		SetType(diskData.Type).
		SetStatus(diskData.Status).
		SetSizeGB(diskData.SizeGb).
		SetArchitecture(diskData.Architecture).
		SetSelfLink(diskData.SelfLink).
		SetCreationTimestamp(diskData.CreationTimestamp).
		SetLastAttachTimestamp(diskData.LastAttachTimestamp).
		SetLastDetachTimestamp(diskData.LastDetachTimestamp).
		SetSourceImage(diskData.SourceImage).
		SetSourceImageID(diskData.SourceImageId).
		SetSourceSnapshot(diskData.SourceSnapshot).
		SetSourceSnapshotID(diskData.SourceSnapshotId).
		SetSourceDisk(diskData.SourceDisk).
		SetSourceDiskID(diskData.SourceDiskId).
		SetProvisionedIops(diskData.ProvisionedIops).
		SetProvisionedThroughput(diskData.ProvisionedThroughput).
		SetPhysicalBlockSizeBytes(diskData.PhysicalBlockSizeBytes).
		SetEnableConfidentialCompute(diskData.EnableConfidentialCompute).
		SetProjectID(diskData.ProjectID)

	if diskData.DiskEncryptionKeyJSON != nil {
		diskHistCreate.SetDiskEncryptionKeyJSON(diskData.DiskEncryptionKeyJSON)
	}
	if diskData.UsersJSON != nil {
		diskHistCreate.SetUsersJSON(diskData.UsersJSON)
	}
	if diskData.ReplicaZonesJSON != nil {
		diskHistCreate.SetReplicaZonesJSON(diskData.ReplicaZonesJSON)
	}
	if diskData.ResourcePoliciesJSON != nil {
		diskHistCreate.SetResourcePoliciesJSON(diskData.ResourcePoliciesJSON)
	}
	if diskData.GuestOsFeaturesJSON != nil {
		diskHistCreate.SetGuestOsFeaturesJSON(diskData.GuestOsFeaturesJSON)
	}

	diskHist, err := diskHistCreate.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create disk history: %w", err)
	}

	// Create children history with disk_history_id
	return h.createChildrenHistory(ctx, tx, diskHist.HistoryID, diskData, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeDisk, new *DiskData, diff *DiskDiff, now time.Time) error {
	// Get current disk history
	currentHist, err := tx.BronzeHistoryGCPComputeDisk.Query().
		Where(
			bronzehistorygcpcomputedisk.ResourceID(old.ID),
			bronzehistorygcpcomputedisk.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current disk history: %w", err)
	}

	// If disk-level fields changed, close old and create new disk history
	if diff.IsChanged {
		// Close old disk history
		if err := tx.BronzeHistoryGCPComputeDisk.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to close disk history: %w", err)
		}

		// Create new disk history
		diskHistCreate := tx.BronzeHistoryGCPComputeDisk.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetDescription(new.Description).
			SetZone(new.Zone).
			SetRegion(new.Region).
			SetType(new.Type).
			SetStatus(new.Status).
			SetSizeGB(new.SizeGb).
			SetArchitecture(new.Architecture).
			SetSelfLink(new.SelfLink).
			SetCreationTimestamp(new.CreationTimestamp).
			SetLastAttachTimestamp(new.LastAttachTimestamp).
			SetLastDetachTimestamp(new.LastDetachTimestamp).
			SetSourceImage(new.SourceImage).
			SetSourceImageID(new.SourceImageId).
			SetSourceSnapshot(new.SourceSnapshot).
			SetSourceSnapshotID(new.SourceSnapshotId).
			SetSourceDisk(new.SourceDisk).
			SetSourceDiskID(new.SourceDiskId).
			SetProvisionedIops(new.ProvisionedIops).
			SetProvisionedThroughput(new.ProvisionedThroughput).
			SetPhysicalBlockSizeBytes(new.PhysicalBlockSizeBytes).
			SetEnableConfidentialCompute(new.EnableConfidentialCompute).
			SetProjectID(new.ProjectID)

		if new.DiskEncryptionKeyJSON != nil {
			diskHistCreate.SetDiskEncryptionKeyJSON(new.DiskEncryptionKeyJSON)
		}
		if new.UsersJSON != nil {
			diskHistCreate.SetUsersJSON(new.UsersJSON)
		}
		if new.ReplicaZonesJSON != nil {
			diskHistCreate.SetReplicaZonesJSON(new.ReplicaZonesJSON)
		}
		if new.ResourcePoliciesJSON != nil {
			diskHistCreate.SetResourcePoliciesJSON(new.ResourcePoliciesJSON)
		}
		if new.GuestOsFeaturesJSON != nil {
			diskHistCreate.SetGuestOsFeaturesJSON(new.GuestOsFeaturesJSON)
		}

		diskHist, err := diskHistCreate.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new disk history: %w", err)
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(ctx, tx, currentHist.HistoryID, now); err != nil {
			return fmt.Errorf("failed to close children history: %w", err)
		}
		return h.createChildrenHistory(ctx, tx, diskHist.HistoryID, new, now)
	}

	// Disk unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(ctx, tx, currentHist.HistoryID, new, diff, now)
}

// CloseHistory closes history records for a deleted disk.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	// Get current disk history
	currentHist, err := tx.BronzeHistoryGCPComputeDisk.Query().
		Where(
			bronzehistorygcpcomputedisk.ResourceID(resourceID),
			bronzehistorygcpcomputedisk.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("failed to find current disk history: %w", err)
	}

	// Close disk history
	if err := tx.BronzeHistoryGCPComputeDisk.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("failed to close disk history: %w", err)
	}

	// Close all children history
	return h.closeChildrenHistory(ctx, tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(ctx context.Context, tx *ent.Tx, diskHistoryID uint, data *DiskData, now time.Time) error {
	// Labels
	for _, labelData := range data.Labels {
		_, err := tx.BronzeHistoryGCPComputeDiskLabel.Create().
			SetDiskHistoryID(diskHistoryID).
			SetValidFrom(now).
			SetKey(labelData.Key).
			SetValue(labelData.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label history: %w", err)
		}
	}

	// Licenses
	for _, licenseData := range data.Licenses {
		_, err := tx.BronzeHistoryGCPComputeDiskLicense.Create().
			SetDiskHistoryID(diskHistoryID).
			SetValidFrom(now).
			SetLicense(licenseData.License).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create license history: %w", err)
		}
	}

	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(ctx context.Context, tx *ent.Tx, diskHistoryID uint, now time.Time) error {
	// Close labels
	_, err := tx.BronzeHistoryGCPComputeDiskLabel.Update().
		Where(
			bronzehistorygcpcomputedisklabel.DiskHistoryID(diskHistoryID),
			bronzehistorygcpcomputedisklabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close labels history: %w", err)
	}

	// Close licenses
	_, err = tx.BronzeHistoryGCPComputeDiskLicense.Update().
		Where(
			bronzehistorygcpcomputedisklicense.DiskHistoryID(diskHistoryID),
			bronzehistorygcpcomputedisklicense.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close licenses history: %w", err)
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(ctx context.Context, tx *ent.Tx, diskHistoryID uint, new *DiskData, diff *DiskDiff, now time.Time) error {
	if diff.LabelsDiff.Changed {
		if err := h.updateLabelsHistory(ctx, tx, diskHistoryID, new.Labels, now); err != nil {
			return err
		}
	}

	if diff.LicensesDiff.Changed {
		if err := h.updateLicensesHistory(ctx, tx, diskHistoryID, new.Licenses, now); err != nil {
			return err
		}
	}

	return nil
}

func (h *HistoryService) updateLabelsHistory(ctx context.Context, tx *ent.Tx, diskHistoryID uint, labels []DiskLabelData, now time.Time) error {
	// Close old labels
	_, err := tx.BronzeHistoryGCPComputeDiskLabel.Update().
		Where(
			bronzehistorygcpcomputedisklabel.DiskHistoryID(diskHistoryID),
			bronzehistorygcpcomputedisklabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close labels history: %w", err)
	}

	// Create new labels
	for _, labelData := range labels {
		_, err := tx.BronzeHistoryGCPComputeDiskLabel.Create().
			SetDiskHistoryID(diskHistoryID).
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

func (h *HistoryService) updateLicensesHistory(ctx context.Context, tx *ent.Tx, diskHistoryID uint, licenses []DiskLicenseData, now time.Time) error {
	// Close old licenses
	_, err := tx.BronzeHistoryGCPComputeDiskLicense.Update().
		Where(
			bronzehistorygcpcomputedisklicense.DiskHistoryID(diskHistoryID),
			bronzehistorygcpcomputedisklicense.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close licenses history: %w", err)
	}

	// Create new licenses
	for _, licenseData := range licenses {
		_, err := tx.BronzeHistoryGCPComputeDiskLicense.Create().
			SetDiskHistoryID(diskHistoryID).
			SetValidFrom(now).
			SetLicense(licenseData.License).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create license history: %w", err)
		}
	}

	return nil
}

