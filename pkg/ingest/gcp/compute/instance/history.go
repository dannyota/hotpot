package instance

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputeinstance"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputeinstancedisk"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputeinstancedisklicense"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputeinstancelabel"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputeinstancemetadata"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputeinstancenic"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputeinstancenicaccessconfig"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputeinstancenicaliasrange"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputeinstanceserviceaccount"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputeinstancetag"
)

// HistoryService handles history tracking for instances.
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
	instHistCreate := tx.BronzeHistoryGCPComputeInstance.Create().
		SetResourceID(instanceData.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(instanceData.CollectedAt).
		SetFirstCollectedAt(instanceData.CollectedAt).
		SetName(instanceData.Name).
		SetZone(instanceData.Zone).
		SetMachineType(instanceData.MachineType).
		SetStatus(instanceData.Status).
		SetStatusMessage(instanceData.StatusMessage).
		SetCPUPlatform(instanceData.CpuPlatform).
		SetHostname(instanceData.Hostname).
		SetDescription(instanceData.Description).
		SetCreationTimestamp(instanceData.CreationTimestamp).
		SetLastStartTimestamp(instanceData.LastStartTimestamp).
		SetLastStopTimestamp(instanceData.LastStopTimestamp).
		SetLastSuspendedTimestamp(instanceData.LastSuspendedTimestamp).
		SetDeletionProtection(instanceData.DeletionProtection).
		SetCanIPForward(instanceData.CanIpForward).
		SetSelfLink(instanceData.SelfLink).
		SetProjectID(instanceData.ProjectID)

	if instanceData.SchedulingJSON != nil {
		instHistCreate.SetSchedulingJSON(instanceData.SchedulingJSON)
	}

	instHist, err := instHistCreate.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create instance history: %w", err)
	}

	// Create children history with instance_history_id
	return h.createChildrenHistory(ctx, tx, instHist.HistoryID, instanceData, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeInstance, new *InstanceData, diff *InstanceDiff, now time.Time) error {
	// Get current instance history
	currentHist, err := tx.BronzeHistoryGCPComputeInstance.Query().
		Where(
			bronzehistorygcpcomputeinstance.ResourceID(old.ID),
			bronzehistorygcpcomputeinstance.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current instance history: %w", err)
	}

	// If instance-level fields changed, close old and create new instance history
	if diff.IsChanged {
		// Close old instance history
		if err := tx.BronzeHistoryGCPComputeInstance.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to close instance history: %w", err)
		}

		// Create new instance history
		instHistCreate := tx.BronzeHistoryGCPComputeInstance.Create().
			SetResourceID(new.ResourceID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetZone(new.Zone).
			SetMachineType(new.MachineType).
			SetStatus(new.Status).
			SetStatusMessage(new.StatusMessage).
			SetCPUPlatform(new.CpuPlatform).
			SetHostname(new.Hostname).
			SetDescription(new.Description).
			SetCreationTimestamp(new.CreationTimestamp).
			SetLastStartTimestamp(new.LastStartTimestamp).
			SetLastStopTimestamp(new.LastStopTimestamp).
			SetLastSuspendedTimestamp(new.LastSuspendedTimestamp).
			SetDeletionProtection(new.DeletionProtection).
			SetCanIPForward(new.CanIpForward).
			SetSelfLink(new.SelfLink).
			SetProjectID(new.ProjectID)

		if new.SchedulingJSON != nil {
			instHistCreate.SetSchedulingJSON(new.SchedulingJSON)
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
	return h.updateChildrenHistory(ctx, tx, currentHist.HistoryID, old, new, diff, now)
}

// CloseHistory closes history records for a deleted instance.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	// Get current instance history
	currentHist, err := tx.BronzeHistoryGCPComputeInstance.Query().
		Where(
			bronzehistorygcpcomputeinstance.ResourceID(resourceID),
			bronzehistorygcpcomputeinstance.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("failed to find current instance history: %w", err)
	}

	// Close instance history
	if err := tx.BronzeHistoryGCPComputeInstance.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("failed to close instance history: %w", err)
	}

	// Close all children history
	return h.closeChildrenHistory(ctx, tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(ctx context.Context, tx *ent.Tx, instanceHistoryID uint, data *InstanceData, now time.Time) error {
	// Disks
	for _, diskData := range data.Disks {
		diskHistCreate := tx.BronzeHistoryGCPComputeInstanceDisk.Create().
			SetInstanceHistoryID(instanceHistoryID).
			SetValidFrom(now).
			SetSource(diskData.Source).
			SetDeviceName(diskData.DeviceName).
			SetIndex(diskData.Index).
			SetBoot(diskData.Boot).
			SetAutoDelete(diskData.AutoDelete).
			SetMode(diskData.Mode).
			SetInterface(diskData.Interface).
			SetType(diskData.Type).
			SetDiskSizeGB(diskData.DiskSizeGb)

		if diskData.DiskEncryptionKeyJSON != nil {
			diskHistCreate.SetDiskEncryptionKeyJSON(diskData.DiskEncryptionKeyJSON)
		}
		if diskData.InitializeParamsJSON != nil {
			diskHistCreate.SetInitializeParamsJSON(diskData.InitializeParamsJSON)
		}

		diskHist, err := diskHistCreate.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create disk history: %w", err)
		}

		// Disk licenses
		for _, licData := range diskData.Licenses {
			_, err := tx.BronzeHistoryGCPComputeInstanceDiskLicense.Create().
				SetDiskHistoryID(diskHist.HistoryID).
				SetValidFrom(now).
				SetLicense(licData.License).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create disk license history: %w", err)
			}
		}
	}

	// NICs
	for _, nicData := range data.NICs {
		nicHistCreate := tx.BronzeHistoryGCPComputeInstanceNIC.Create().
			SetInstanceHistoryID(instanceHistoryID).
			SetValidFrom(now).
			SetName(nicData.Name).
			SetNetwork(nicData.Network).
			SetSubnetwork(nicData.Subnetwork).
			SetNetworkIP(nicData.NetworkIP).
			SetStackType(nicData.StackType).
			SetNicType(nicData.NicType)

		nicHist, err := nicHistCreate.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create NIC history: %w", err)
		}

		// Access configs
		for _, acData := range nicData.AccessConfigs {
			_, err := tx.BronzeHistoryGCPComputeInstanceNICAccessConfig.Create().
				SetNicHistoryID(nicHist.HistoryID).
				SetValidFrom(now).
				SetType(acData.Type).
				SetName(acData.Name).
				SetNatIP(acData.NatIP).
				SetNetworkTier(acData.NetworkTier).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create access config history: %w", err)
			}
		}

		// Alias ranges
		for _, arData := range nicData.AliasIPRanges {
			_, err := tx.BronzeHistoryGCPComputeInstanceNICAliasRange.Create().
				SetNicHistoryID(nicHist.HistoryID).
				SetValidFrom(now).
				SetIPCidrRange(arData.IPCidrRange).
				SetSubnetworkRangeName(arData.SubnetworkRangeName).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create alias range history: %w", err)
			}
		}
	}

	// Labels
	for _, labelData := range data.Labels {
		_, err := tx.BronzeHistoryGCPComputeInstanceLabel.Create().
			SetInstanceHistoryID(instanceHistoryID).
			SetValidFrom(now).
			SetKey(labelData.Key).
			SetValue(labelData.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label history: %w", err)
		}
	}

	// Tags
	for _, tagData := range data.Tags {
		_, err := tx.BronzeHistoryGCPComputeInstanceTag.Create().
			SetInstanceHistoryID(instanceHistoryID).
			SetValidFrom(now).
			SetTag(tagData.Tag).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create tag history: %w", err)
		}
	}

	// Metadata
	for _, metaData := range data.Metadata {
		_, err := tx.BronzeHistoryGCPComputeInstanceMetadata.Create().
			SetInstanceHistoryID(instanceHistoryID).
			SetValidFrom(now).
			SetKey(metaData.Key).
			SetValue(metaData.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create metadata history: %w", err)
		}
	}

	// Service accounts
	for _, saData := range data.ServiceAccounts {
		saCreate := tx.BronzeHistoryGCPComputeInstanceServiceAccount.Create().
			SetInstanceHistoryID(instanceHistoryID).
			SetValidFrom(now).
			SetEmail(saData.Email)

		if saData.ScopesJSON != nil {
			saCreate.SetScopesJSON(saData.ScopesJSON)
		}

		_, err := saCreate.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create service account history: %w", err)
		}
	}

	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(ctx context.Context, tx *ent.Tx, instanceHistoryID uint, now time.Time) error {
	// Close direct children (disks, NICs, labels, tags, metadata, service accounts)
	_, err := tx.BronzeHistoryGCPComputeInstanceDisk.Update().
		Where(
			bronzehistorygcpcomputeinstancedisk.InstanceHistoryID(instanceHistoryID),
			bronzehistorygcpcomputeinstancedisk.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close disk history: %w", err)
	}

	_, err = tx.BronzeHistoryGCPComputeInstanceNIC.Update().
		Where(
			bronzehistorygcpcomputeinstancenic.InstanceHistoryID(instanceHistoryID),
			bronzehistorygcpcomputeinstancenic.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close NIC history: %w", err)
	}

	_, err = tx.BronzeHistoryGCPComputeInstanceLabel.Update().
		Where(
			bronzehistorygcpcomputeinstancelabel.InstanceHistoryID(instanceHistoryID),
			bronzehistorygcpcomputeinstancelabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close label history: %w", err)
	}

	_, err = tx.BronzeHistoryGCPComputeInstanceTag.Update().
		Where(
			bronzehistorygcpcomputeinstancetag.InstanceHistoryID(instanceHistoryID),
			bronzehistorygcpcomputeinstancetag.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close tag history: %w", err)
	}

	_, err = tx.BronzeHistoryGCPComputeInstanceMetadata.Update().
		Where(
			bronzehistorygcpcomputeinstancemetadata.InstanceHistoryID(instanceHistoryID),
			bronzehistorygcpcomputeinstancemetadata.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close metadata history: %w", err)
	}

	_, err = tx.BronzeHistoryGCPComputeInstanceServiceAccount.Update().
		Where(
			bronzehistorygcpcomputeinstanceserviceaccount.InstanceHistoryID(instanceHistoryID),
			bronzehistorygcpcomputeinstanceserviceaccount.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close service account history: %w", err)
	}

	// Close nested children (disk licenses, NIC access configs, NIC alias ranges)
	// Get disk history IDs
	var diskHistIDs []uint
	err = tx.BronzeHistoryGCPComputeInstanceDisk.Query().
		Where(bronzehistorygcpcomputeinstancedisk.InstanceHistoryID(instanceHistoryID)).
		Select(bronzehistorygcpcomputeinstancedisk.FieldHistoryID).
		Scan(ctx, &diskHistIDs)
	if err != nil {
		return fmt.Errorf("failed to get disk history IDs: %w", err)
	}

	if len(diskHistIDs) > 0 {
		_, err = tx.BronzeHistoryGCPComputeInstanceDiskLicense.Update().
			Where(
				bronzehistorygcpcomputeinstancedisklicense.DiskHistoryIDIn(diskHistIDs...),
				bronzehistorygcpcomputeinstancedisklicense.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to close disk license history: %w", err)
		}
	}

	// Get NIC history IDs
	var nicHistIDs []uint
	err = tx.BronzeHistoryGCPComputeInstanceNIC.Query().
		Where(bronzehistorygcpcomputeinstancenic.InstanceHistoryID(instanceHistoryID)).
		Select(bronzehistorygcpcomputeinstancenic.FieldHistoryID).
		Scan(ctx, &nicHistIDs)
	if err != nil {
		return fmt.Errorf("failed to get NIC history IDs: %w", err)
	}

	if len(nicHistIDs) > 0 {
		_, err = tx.BronzeHistoryGCPComputeInstanceNICAccessConfig.Update().
			Where(
				bronzehistorygcpcomputeinstancenicaccessconfig.NicHistoryIDIn(nicHistIDs...),
				bronzehistorygcpcomputeinstancenicaccessconfig.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to close NIC access config history: %w", err)
		}

		_, err = tx.BronzeHistoryGCPComputeInstanceNICAliasRange.Update().
			Where(
				bronzehistorygcpcomputeinstancenicaliasrange.NicHistoryIDIn(nicHistIDs...),
				bronzehistorygcpcomputeinstancenicaliasrange.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to close NIC alias range history: %w", err)
		}
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(ctx context.Context, tx *ent.Tx, instanceHistoryID uint, old *ent.BronzeGCPComputeInstance, new *InstanceData, diff *InstanceDiff, now time.Time) error {
	// For each child type, if changed: close old + create new
	// If unchanged: no action (still links to same instance_history_id)

	if diff.DisksDiff.Changed {
		if err := h.updateDisksHistory(ctx, tx, instanceHistoryID, new.Disks, now); err != nil {
			return err
		}
	}

	if diff.NICsDiff.Changed {
		if err := h.updateNICsHistory(ctx, tx, instanceHistoryID, new.NICs, now); err != nil {
			return err
		}
	}

	if diff.LabelsDiff.Changed {
		if err := h.updateLabelsHistory(ctx, tx, instanceHistoryID, new.Labels, now); err != nil {
			return err
		}
	}

	if diff.TagsDiff.Changed {
		if err := h.updateTagsHistory(ctx, tx, instanceHistoryID, new.Tags, now); err != nil {
			return err
		}
	}

	if diff.MetadataDiff.Changed {
		if err := h.updateMetadataHistory(ctx, tx, instanceHistoryID, new.Metadata, now); err != nil {
			return err
		}
	}

	if diff.ServiceAccountsDiff.Changed {
		if err := h.updateServiceAccountsHistory(ctx, tx, instanceHistoryID, new.ServiceAccounts, now); err != nil {
			return err
		}
	}

	return nil
}

func (h *HistoryService) updateDisksHistory(ctx context.Context, tx *ent.Tx, instanceHistoryID uint, disks []DiskData, now time.Time) error {
	// Get old disk history IDs to close nested licenses
	var oldDiskHistIDs []uint
	err := tx.BronzeHistoryGCPComputeInstanceDisk.Query().
		Where(
			bronzehistorygcpcomputeinstancedisk.InstanceHistoryID(instanceHistoryID),
			bronzehistorygcpcomputeinstancedisk.ValidToIsNil(),
		).
		Select(bronzehistorygcpcomputeinstancedisk.FieldHistoryID).
		Scan(ctx, &oldDiskHistIDs)
	if err != nil {
		return fmt.Errorf("failed to get old disk history IDs: %w", err)
	}

	// Close old disk licenses
	if len(oldDiskHistIDs) > 0 {
		_, err = tx.BronzeHistoryGCPComputeInstanceDiskLicense.Update().
			Where(
				bronzehistorygcpcomputeinstancedisklicense.DiskHistoryIDIn(oldDiskHistIDs...),
				bronzehistorygcpcomputeinstancedisklicense.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to close disk license history: %w", err)
		}
	}

	// Close old disk history
	_, err = tx.BronzeHistoryGCPComputeInstanceDisk.Update().
		Where(
			bronzehistorygcpcomputeinstancedisk.InstanceHistoryID(instanceHistoryID),
			bronzehistorygcpcomputeinstancedisk.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close disk history: %w", err)
	}

	// Create new disk history
	for _, diskData := range disks {
		diskHistCreate := tx.BronzeHistoryGCPComputeInstanceDisk.Create().
			SetInstanceHistoryID(instanceHistoryID).
			SetValidFrom(now).
			SetSource(diskData.Source).
			SetDeviceName(diskData.DeviceName).
			SetIndex(diskData.Index).
			SetBoot(diskData.Boot).
			SetAutoDelete(diskData.AutoDelete).
			SetMode(diskData.Mode).
			SetInterface(diskData.Interface).
			SetType(diskData.Type).
			SetDiskSizeGB(diskData.DiskSizeGb)

		if diskData.DiskEncryptionKeyJSON != nil {
			diskHistCreate.SetDiskEncryptionKeyJSON(diskData.DiskEncryptionKeyJSON)
		}
		if diskData.InitializeParamsJSON != nil {
			diskHistCreate.SetInitializeParamsJSON(diskData.InitializeParamsJSON)
		}

		diskHist, err := diskHistCreate.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create disk history: %w", err)
		}

		for _, licData := range diskData.Licenses {
			_, err := tx.BronzeHistoryGCPComputeInstanceDiskLicense.Create().
				SetDiskHistoryID(diskHist.HistoryID).
				SetValidFrom(now).
				SetLicense(licData.License).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create disk license history: %w", err)
			}
		}
	}

	return nil
}

func (h *HistoryService) updateNICsHistory(ctx context.Context, tx *ent.Tx, instanceHistoryID uint, nics []NICData, now time.Time) error {
	// Get old NIC history IDs to close nested children
	var oldNICHistIDs []uint
	err := tx.BronzeHistoryGCPComputeInstanceNIC.Query().
		Where(
			bronzehistorygcpcomputeinstancenic.InstanceHistoryID(instanceHistoryID),
			bronzehistorygcpcomputeinstancenic.ValidToIsNil(),
		).
		Select(bronzehistorygcpcomputeinstancenic.FieldHistoryID).
		Scan(ctx, &oldNICHistIDs)
	if err != nil {
		return fmt.Errorf("failed to get old NIC history IDs: %w", err)
	}

	// Close old NIC access configs and alias ranges
	if len(oldNICHistIDs) > 0 {
		_, err = tx.BronzeHistoryGCPComputeInstanceNICAccessConfig.Update().
			Where(
				bronzehistorygcpcomputeinstancenicaccessconfig.NicHistoryIDIn(oldNICHistIDs...),
				bronzehistorygcpcomputeinstancenicaccessconfig.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to close access config history: %w", err)
		}

		_, err = tx.BronzeHistoryGCPComputeInstanceNICAliasRange.Update().
			Where(
				bronzehistorygcpcomputeinstancenicaliasrange.NicHistoryIDIn(oldNICHistIDs...),
				bronzehistorygcpcomputeinstancenicaliasrange.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to close alias range history: %w", err)
		}
	}

	// Close old NIC history
	_, err = tx.BronzeHistoryGCPComputeInstanceNIC.Update().
		Where(
			bronzehistorygcpcomputeinstancenic.InstanceHistoryID(instanceHistoryID),
			bronzehistorygcpcomputeinstancenic.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close NIC history: %w", err)
	}

	// Create new NIC history
	for _, nicData := range nics {
		nicHist, err := tx.BronzeHistoryGCPComputeInstanceNIC.Create().
			SetInstanceHistoryID(instanceHistoryID).
			SetValidFrom(now).
			SetName(nicData.Name).
			SetNetwork(nicData.Network).
			SetSubnetwork(nicData.Subnetwork).
			SetNetworkIP(nicData.NetworkIP).
			SetStackType(nicData.StackType).
			SetNicType(nicData.NicType).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create NIC history: %w", err)
		}

		for _, acData := range nicData.AccessConfigs {
			_, err := tx.BronzeHistoryGCPComputeInstanceNICAccessConfig.Create().
				SetNicHistoryID(nicHist.HistoryID).
				SetValidFrom(now).
				SetType(acData.Type).
				SetName(acData.Name).
				SetNatIP(acData.NatIP).
				SetNetworkTier(acData.NetworkTier).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create access config history: %w", err)
			}
		}

		for _, arData := range nicData.AliasIPRanges {
			_, err := tx.BronzeHistoryGCPComputeInstanceNICAliasRange.Create().
				SetNicHistoryID(nicHist.HistoryID).
				SetValidFrom(now).
				SetIPCidrRange(arData.IPCidrRange).
				SetSubnetworkRangeName(arData.SubnetworkRangeName).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create alias range history: %w", err)
			}
		}
	}

	return nil
}

func (h *HistoryService) updateLabelsHistory(ctx context.Context, tx *ent.Tx, instanceHistoryID uint, labels []LabelData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPComputeInstanceLabel.Update().
		Where(
			bronzehistorygcpcomputeinstancelabel.InstanceHistoryID(instanceHistoryID),
			bronzehistorygcpcomputeinstancelabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close label history: %w", err)
	}

	for _, labelData := range labels {
		_, err := tx.BronzeHistoryGCPComputeInstanceLabel.Create().
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

func (h *HistoryService) updateTagsHistory(ctx context.Context, tx *ent.Tx, instanceHistoryID uint, tags []TagData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPComputeInstanceTag.Update().
		Where(
			bronzehistorygcpcomputeinstancetag.InstanceHistoryID(instanceHistoryID),
			bronzehistorygcpcomputeinstancetag.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close tag history: %w", err)
	}

	for _, tagData := range tags {
		_, err := tx.BronzeHistoryGCPComputeInstanceTag.Create().
			SetInstanceHistoryID(instanceHistoryID).
			SetValidFrom(now).
			SetTag(tagData.Tag).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create tag history: %w", err)
		}
	}
	return nil
}

func (h *HistoryService) updateMetadataHistory(ctx context.Context, tx *ent.Tx, instanceHistoryID uint, metadata []MetadataData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPComputeInstanceMetadata.Update().
		Where(
			bronzehistorygcpcomputeinstancemetadata.InstanceHistoryID(instanceHistoryID),
			bronzehistorygcpcomputeinstancemetadata.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close metadata history: %w", err)
	}

	for _, metaData := range metadata {
		_, err := tx.BronzeHistoryGCPComputeInstanceMetadata.Create().
			SetInstanceHistoryID(instanceHistoryID).
			SetValidFrom(now).
			SetKey(metaData.Key).
			SetValue(metaData.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create metadata history: %w", err)
		}
	}
	return nil
}

func (h *HistoryService) updateServiceAccountsHistory(ctx context.Context, tx *ent.Tx, instanceHistoryID uint, sas []ServiceAccountData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPComputeInstanceServiceAccount.Update().
		Where(
			bronzehistorygcpcomputeinstanceserviceaccount.InstanceHistoryID(instanceHistoryID),
			bronzehistorygcpcomputeinstanceserviceaccount.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close service account history: %w", err)
	}

	for _, saData := range sas {
		saCreate := tx.BronzeHistoryGCPComputeInstanceServiceAccount.Create().
			SetInstanceHistoryID(instanceHistoryID).
			SetValidFrom(now).
			SetEmail(saData.Email)

		if saData.ScopesJSON != nil {
			saCreate.SetScopesJSON(saData.ScopesJSON)
		}

		_, err := saCreate.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create service account history: %w", err)
		}
	}
	return nil
}
