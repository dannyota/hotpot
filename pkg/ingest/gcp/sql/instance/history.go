package instance

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpsqlinstance"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpsqlinstancelabel"
)

// HistoryService handles history tracking for SQL instances.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new instance and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, instanceData *InstanceData, now time.Time) error {
	// Create instance history
	instHistCreate := tx.BronzeHistoryGCPSQLInstance.Create().
		SetResourceID(instanceData.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(instanceData.CollectedAt).
		SetFirstCollectedAt(instanceData.CollectedAt).
		SetName(instanceData.Name).
		SetDatabaseVersion(instanceData.DatabaseVersion).
		SetState(instanceData.State).
		SetRegion(instanceData.Region).
		SetGceZone(instanceData.GceZone).
		SetSecondaryGceZone(instanceData.SecondaryGceZone).
		SetInstanceType(instanceData.InstanceType).
		SetConnectionName(instanceData.ConnectionName).
		SetServiceAccountEmailAddress(instanceData.ServiceAccountEmailAddress).
		SetSelfLink(instanceData.SelfLink).
		SetProjectID(instanceData.ProjectID)

	if instanceData.SettingsJSON != nil {
		instHistCreate.SetSettingsJSON(instanceData.SettingsJSON)
	}
	if instanceData.ServerCaCertJSON != nil {
		instHistCreate.SetServerCaCertJSON(instanceData.ServerCaCertJSON)
	}
	if instanceData.IpAddressesJSON != nil {
		instHistCreate.SetIPAddressesJSON(instanceData.IpAddressesJSON)
	}
	if instanceData.ReplicaConfigurationJSON != nil {
		instHistCreate.SetReplicaConfigurationJSON(instanceData.ReplicaConfigurationJSON)
	}
	if instanceData.FailoverReplicaJSON != nil {
		instHistCreate.SetFailoverReplicaJSON(instanceData.FailoverReplicaJSON)
	}
	if instanceData.DiskEncryptionConfigurationJSON != nil {
		instHistCreate.SetDiskEncryptionConfigurationJSON(instanceData.DiskEncryptionConfigurationJSON)
	}
	if instanceData.DiskEncryptionStatusJSON != nil {
		instHistCreate.SetDiskEncryptionStatusJSON(instanceData.DiskEncryptionStatusJSON)
	}

	instHist, err := instHistCreate.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create instance history: %w", err)
	}

	// Create children history with instance_history_id
	return h.createChildrenHistory(ctx, tx, instHist.HistoryID, instanceData, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPSQLInstance, new *InstanceData, diff *InstanceDiff, now time.Time) error {
	// Get current instance history
	currentHist, err := tx.BronzeHistoryGCPSQLInstance.Query().
		Where(
			bronzehistorygcpsqlinstance.ResourceID(old.ID),
			bronzehistorygcpsqlinstance.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current instance history: %w", err)
	}

	// If instance-level fields changed, close old and create new instance history
	if diff.IsChanged {
		// Close old instance history
		if err := tx.BronzeHistoryGCPSQLInstance.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to close instance history: %w", err)
		}

		// Create new instance history
		instHistCreate := tx.BronzeHistoryGCPSQLInstance.Create().
			SetResourceID(new.ResourceID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetDatabaseVersion(new.DatabaseVersion).
			SetState(new.State).
			SetRegion(new.Region).
			SetGceZone(new.GceZone).
			SetSecondaryGceZone(new.SecondaryGceZone).
			SetInstanceType(new.InstanceType).
			SetConnectionName(new.ConnectionName).
			SetServiceAccountEmailAddress(new.ServiceAccountEmailAddress).
			SetSelfLink(new.SelfLink).
			SetProjectID(new.ProjectID)

		if new.SettingsJSON != nil {
			instHistCreate.SetSettingsJSON(new.SettingsJSON)
		}
		if new.ServerCaCertJSON != nil {
			instHistCreate.SetServerCaCertJSON(new.ServerCaCertJSON)
		}
		if new.IpAddressesJSON != nil {
			instHistCreate.SetIPAddressesJSON(new.IpAddressesJSON)
		}
		if new.ReplicaConfigurationJSON != nil {
			instHistCreate.SetReplicaConfigurationJSON(new.ReplicaConfigurationJSON)
		}
		if new.FailoverReplicaJSON != nil {
			instHistCreate.SetFailoverReplicaJSON(new.FailoverReplicaJSON)
		}
		if new.DiskEncryptionConfigurationJSON != nil {
			instHistCreate.SetDiskEncryptionConfigurationJSON(new.DiskEncryptionConfigurationJSON)
		}
		if new.DiskEncryptionStatusJSON != nil {
			instHistCreate.SetDiskEncryptionStatusJSON(new.DiskEncryptionStatusJSON)
		}

		instHist, err := instHistCreate.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new instance history: %w", err)
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(ctx, tx, currentHist.HistoryID, now); err != nil {
			return fmt.Errorf("failed to close children history: %w", err)
		}
		return h.createChildrenHistory(ctx, tx, instHist.HistoryID, new, now)
	}

	// Instance unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(ctx, tx, currentHist.HistoryID, new, diff, now)
}

// CloseHistory closes history records for a deleted instance.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	// Get current instance history
	currentHist, err := tx.BronzeHistoryGCPSQLInstance.Query().
		Where(
			bronzehistorygcpsqlinstance.ResourceID(resourceID),
			bronzehistorygcpsqlinstance.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("failed to find current instance history: %w", err)
	}

	// Close instance history
	if err := tx.BronzeHistoryGCPSQLInstance.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("failed to close instance history: %w", err)
	}

	// Close all children history
	return h.closeChildrenHistory(ctx, tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(ctx context.Context, tx *ent.Tx, instanceHistoryID uint, data *InstanceData, now time.Time) error {
	// Labels
	for _, labelData := range data.Labels {
		_, err := tx.BronzeHistoryGCPSQLInstanceLabel.Create().
			SetInstanceHistoryID(instanceHistoryID).
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
func (h *HistoryService) closeChildrenHistory(ctx context.Context, tx *ent.Tx, instanceHistoryID uint, now time.Time) error {
	// Close labels
	_, err := tx.BronzeHistoryGCPSQLInstanceLabel.Update().
		Where(
			bronzehistorygcpsqlinstancelabel.InstanceHistoryID(instanceHistoryID),
			bronzehistorygcpsqlinstancelabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close label history: %w", err)
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(ctx context.Context, tx *ent.Tx, instanceHistoryID uint, new *InstanceData, diff *InstanceDiff, now time.Time) error {
	if diff.LabelsDiff.Changed {
		if err := h.updateLabelsHistory(ctx, tx, instanceHistoryID, new.Labels, now); err != nil {
			return err
		}
	}

	return nil
}

func (h *HistoryService) updateLabelsHistory(ctx context.Context, tx *ent.Tx, instanceHistoryID uint, labels []LabelData, now time.Time) error {
	// Close old label history
	_, err := tx.BronzeHistoryGCPSQLInstanceLabel.Update().
		Where(
			bronzehistorygcpsqlinstancelabel.InstanceHistoryID(instanceHistoryID),
			bronzehistorygcpsqlinstancelabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close label history: %w", err)
	}

	// Create new label history
	for _, labelData := range labels {
		_, err := tx.BronzeHistoryGCPSQLInstanceLabel.Create().
			SetInstanceHistoryID(instanceHistoryID).
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
