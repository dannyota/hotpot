package snapshot

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcomputesnapshot"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcomputesnapshotlabel"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcomputesnapshotlicense"
)

// HistoryService handles history tracking for snapshots.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new snapshot and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, snapshotData *SnapshotData, now time.Time) error {
	// Create snapshot history
	snapHistCreate := tx.BronzeHistoryGCPComputeSnapshot.Create().
		SetResourceID(snapshotData.ID).
		SetValidFrom(now).
		SetCollectedAt(snapshotData.CollectedAt).
		SetFirstCollectedAt(snapshotData.CollectedAt).
		SetName(snapshotData.Name).
		SetDescription(snapshotData.Description).
		SetStatus(snapshotData.Status).
		SetDiskSizeGB(snapshotData.DiskSizeGB).
		SetStorageBytes(snapshotData.StorageBytes).
		SetStorageBytesStatus(snapshotData.StorageBytesStatus).
		SetDownloadBytes(snapshotData.DownloadBytes).
		SetSnapshotType(snapshotData.SnapshotType).
		SetArchitecture(snapshotData.Architecture).
		SetSelfLink(snapshotData.SelfLink).
		SetCreationTimestamp(snapshotData.CreationTimestamp).
		SetLabelFingerprint(snapshotData.LabelFingerprint).
		SetSourceDisk(snapshotData.SourceDisk).
		SetSourceDiskID(snapshotData.SourceDiskID).
		SetSourceDiskForRecoveryCheckpoint(snapshotData.SourceDiskForRecoveryCheckpoint).
		SetAutoCreated(snapshotData.AutoCreated).
		SetSatisfiesPzi(snapshotData.SatisfiesPzi).
		SetSatisfiesPzs(snapshotData.SatisfiesPzs).
		SetEnableConfidentialCompute(snapshotData.EnableConfidentialCompute).
		SetProjectID(snapshotData.ProjectID)

	if snapshotData.SnapshotEncryptionKeyJSON != nil {
		snapHistCreate.SetSnapshotEncryptionKeyJSON(snapshotData.SnapshotEncryptionKeyJSON)
	}
	if snapshotData.SourceDiskEncryptionKeyJSON != nil {
		snapHistCreate.SetSourceDiskEncryptionKeyJSON(snapshotData.SourceDiskEncryptionKeyJSON)
	}
	if snapshotData.GuestOsFeaturesJSON != nil {
		snapHistCreate.SetGuestOsFeaturesJSON(snapshotData.GuestOsFeaturesJSON)
	}
	if snapshotData.StorageLocationsJSON != nil {
		snapHistCreate.SetStorageLocationsJSON(snapshotData.StorageLocationsJSON)
	}

	snapHist, err := snapHistCreate.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create snapshot history: %w", err)
	}

	// Create children history with snapshot_history_id
	return h.createChildrenHistory(ctx, tx, snapHist.HistoryID, snapshotData, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeSnapshot, new *SnapshotData, diff *SnapshotDiff, now time.Time) error {
	// Get current snapshot history
	currentHist, err := tx.BronzeHistoryGCPComputeSnapshot.Query().
		Where(
			bronzehistorygcpcomputesnapshot.ResourceID(old.ID),
			bronzehistorygcpcomputesnapshot.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current snapshot history: %w", err)
	}

	// If snapshot-level fields changed, close old and create new snapshot history
	if diff.IsChanged {
		// Close old snapshot history
		if err := tx.BronzeHistoryGCPComputeSnapshot.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to close snapshot history: %w", err)
		}

		// Create new snapshot history
		snapHistCreate := tx.BronzeHistoryGCPComputeSnapshot.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetDescription(new.Description).
			SetStatus(new.Status).
			SetDiskSizeGB(new.DiskSizeGB).
			SetStorageBytes(new.StorageBytes).
			SetStorageBytesStatus(new.StorageBytesStatus).
			SetDownloadBytes(new.DownloadBytes).
			SetSnapshotType(new.SnapshotType).
			SetArchitecture(new.Architecture).
			SetSelfLink(new.SelfLink).
			SetCreationTimestamp(new.CreationTimestamp).
			SetLabelFingerprint(new.LabelFingerprint).
			SetSourceDisk(new.SourceDisk).
			SetSourceDiskID(new.SourceDiskID).
			SetSourceDiskForRecoveryCheckpoint(new.SourceDiskForRecoveryCheckpoint).
			SetAutoCreated(new.AutoCreated).
			SetSatisfiesPzi(new.SatisfiesPzi).
			SetSatisfiesPzs(new.SatisfiesPzs).
			SetEnableConfidentialCompute(new.EnableConfidentialCompute).
			SetProjectID(new.ProjectID)

		if new.SnapshotEncryptionKeyJSON != nil {
			snapHistCreate.SetSnapshotEncryptionKeyJSON(new.SnapshotEncryptionKeyJSON)
		}
		if new.SourceDiskEncryptionKeyJSON != nil {
			snapHistCreate.SetSourceDiskEncryptionKeyJSON(new.SourceDiskEncryptionKeyJSON)
		}
		if new.GuestOsFeaturesJSON != nil {
			snapHistCreate.SetGuestOsFeaturesJSON(new.GuestOsFeaturesJSON)
		}
		if new.StorageLocationsJSON != nil {
			snapHistCreate.SetStorageLocationsJSON(new.StorageLocationsJSON)
		}

		snapHist, err := snapHistCreate.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new snapshot history: %w", err)
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(ctx, tx, currentHist.HistoryID, now); err != nil {
			return fmt.Errorf("failed to close children history: %w", err)
		}
		return h.createChildrenHistory(ctx, tx, snapHist.HistoryID, new, now)
	}

	// Snapshot unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(ctx, tx, currentHist.HistoryID, new, diff, now)
}

// CloseHistory closes history records for a deleted snapshot.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	// Get current snapshot history
	currentHist, err := tx.BronzeHistoryGCPComputeSnapshot.Query().
		Where(
			bronzehistorygcpcomputesnapshot.ResourceID(resourceID),
			bronzehistorygcpcomputesnapshot.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("failed to find current snapshot history: %w", err)
	}

	// Close snapshot history
	if err := tx.BronzeHistoryGCPComputeSnapshot.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("failed to close snapshot history: %w", err)
	}

	// Close all children history
	return h.closeChildrenHistory(ctx, tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(ctx context.Context, tx *ent.Tx, snapshotHistoryID uint, data *SnapshotData, now time.Time) error {
	// Labels
	for _, labelData := range data.Labels {
		_, err := tx.BronzeHistoryGCPComputeSnapshotLabel.Create().
			SetSnapshotHistoryID(snapshotHistoryID).
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
		_, err := tx.BronzeHistoryGCPComputeSnapshotLicense.Create().
			SetSnapshotHistoryID(snapshotHistoryID).
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
func (h *HistoryService) closeChildrenHistory(ctx context.Context, tx *ent.Tx, snapshotHistoryID uint, now time.Time) error {
	// Close labels
	_, err := tx.BronzeHistoryGCPComputeSnapshotLabel.Update().
		Where(
			bronzehistorygcpcomputesnapshotlabel.SnapshotHistoryID(snapshotHistoryID),
			bronzehistorygcpcomputesnapshotlabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close labels history: %w", err)
	}

	// Close licenses
	_, err = tx.BronzeHistoryGCPComputeSnapshotLicense.Update().
		Where(
			bronzehistorygcpcomputesnapshotlicense.SnapshotHistoryID(snapshotHistoryID),
			bronzehistorygcpcomputesnapshotlicense.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close licenses history: %w", err)
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(ctx context.Context, tx *ent.Tx, snapshotHistoryID uint, new *SnapshotData, diff *SnapshotDiff, now time.Time) error {
	if diff.LabelsDiff.Changed {
		if err := h.updateLabelsHistory(ctx, tx, snapshotHistoryID, new.Labels, now); err != nil {
			return err
		}
	}

	if diff.LicensesDiff.Changed {
		if err := h.updateLicensesHistory(ctx, tx, snapshotHistoryID, new.Licenses, now); err != nil {
			return err
		}
	}

	return nil
}

func (h *HistoryService) updateLabelsHistory(ctx context.Context, tx *ent.Tx, snapshotHistoryID uint, labels []SnapshotLabelData, now time.Time) error {
	// Close old labels
	_, err := tx.BronzeHistoryGCPComputeSnapshotLabel.Update().
		Where(
			bronzehistorygcpcomputesnapshotlabel.SnapshotHistoryID(snapshotHistoryID),
			bronzehistorygcpcomputesnapshotlabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close labels history: %w", err)
	}

	// Create new labels
	for _, labelData := range labels {
		_, err := tx.BronzeHistoryGCPComputeSnapshotLabel.Create().
			SetSnapshotHistoryID(snapshotHistoryID).
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

func (h *HistoryService) updateLicensesHistory(ctx context.Context, tx *ent.Tx, snapshotHistoryID uint, licenses []SnapshotLicenseData, now time.Time) error {
	// Close old licenses
	_, err := tx.BronzeHistoryGCPComputeSnapshotLicense.Update().
		Where(
			bronzehistorygcpcomputesnapshotlicense.SnapshotHistoryID(snapshotHistoryID),
			bronzehistorygcpcomputesnapshotlicense.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close licenses history: %w", err)
	}

	// Create new licenses
	for _, licenseData := range licenses {
		_, err := tx.BronzeHistoryGCPComputeSnapshotLicense.Create().
			SetSnapshotHistoryID(snapshotHistoryID).
			SetValidFrom(now).
			SetLicense(licenseData.License).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create license history: %w", err)
		}
	}

	return nil
}
