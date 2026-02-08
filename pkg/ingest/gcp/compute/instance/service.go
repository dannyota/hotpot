package instance

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzegcpcomputeinstance"
	"hotpot/pkg/storage/ent/bronzegcpcomputeinstancedisk"
	"hotpot/pkg/storage/ent/bronzegcpcomputeinstancelabel"
	"hotpot/pkg/storage/ent/bronzegcpcomputeinstancemetadata"
	"hotpot/pkg/storage/ent/bronzegcpcomputeinstancenic"
	"hotpot/pkg/storage/ent/bronzegcpcomputeinstanceserviceaccount"
	"hotpot/pkg/storage/ent/bronzegcpcomputeinstancetag"
)

// Service handles GCP Compute instance ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new instance ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for instance ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of instance ingestion.
type IngestResult struct {
	ProjectID      string
	InstanceCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches instances from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch instances from GCP
	instances, err := s.client.ListInstances(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	// Convert to data structs
	instanceDataList := make([]*InstanceData, 0, len(instances))
	for _, inst := range instances {
		data, err := ConvertInstance(inst, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert instance: %w", err)
		}
		instanceDataList = append(instanceDataList, data)
	}

	// Save to database
	if err := s.saveInstances(ctx, instanceDataList); err != nil {
		return nil, fmt.Errorf("failed to save instances: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		InstanceCount:  len(instanceDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveInstances saves instances to the database with history tracking.
func (s *Service) saveInstances(ctx context.Context, instances []*InstanceData) error {
	if len(instances) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, instanceData := range instances {
		// Load existing instance with all nested edges
		existing, err := tx.BronzeGCPComputeInstance.Query().
			Where(bronzegcpcomputeinstance.ID(instanceData.ResourceID)).
			WithDisks(func(q *ent.BronzeGCPComputeInstanceDiskQuery) {
				q.WithLicenses()
			}).
			WithNics(func(q *ent.BronzeGCPComputeInstanceNICQuery) {
				q.WithAccessConfigs().WithAliasIPRanges()
			}).
			WithLabels().
			WithTags().
			WithMetadata().
			WithServiceAccounts().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing instance %s: %w", instanceData.Name, err)
		}

		// Compute diff
		diff := DiffInstanceData(existing, instanceData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPComputeInstance.UpdateOneID(instanceData.ResourceID).
				SetCollectedAt(instanceData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for instance %s: %w", instanceData.Name, err)
			}
			continue
		}

		// Delete old child entities if updating
		if existing != nil {
			if err := s.deleteInstanceChildren(ctx, tx, instanceData.ResourceID); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old children for instance %s: %w", instanceData.Name, err)
			}
		}

		// Create or update instance
		var savedInstance *ent.BronzeGCPComputeInstance
		if existing == nil {
			// Create new instance
			create := tx.BronzeGCPComputeInstance.Create().
				SetID(instanceData.ResourceID).
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
				SetProjectID(instanceData.ProjectID).
				SetCollectedAt(instanceData.CollectedAt).
				SetFirstCollectedAt(instanceData.CollectedAt)

			if instanceData.SchedulingJSON != nil {
				create.SetSchedulingJSON(instanceData.SchedulingJSON)
			}

			savedInstance, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create instance %s: %w", instanceData.Name, err)
			}
		} else {
			// Update existing instance
			update := tx.BronzeGCPComputeInstance.UpdateOneID(instanceData.ResourceID).
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
				SetProjectID(instanceData.ProjectID).
				SetCollectedAt(instanceData.CollectedAt)

			if instanceData.SchedulingJSON != nil {
				update.SetSchedulingJSON(instanceData.SchedulingJSON)
			}

			savedInstance, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update instance %s: %w", instanceData.Name, err)
			}
		}

		// Create child entities
		if err := s.createInstanceChildren(ctx, tx, savedInstance, instanceData); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create children for instance %s: %w", instanceData.Name, err)
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, instanceData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for instance %s: %w", instanceData.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, instanceData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for instance %s: %w", instanceData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// deleteInstanceChildren deletes all child entities for an instance.
// Note: Ent CASCADE DELETE is set on edges, so deleting parent entities
// will automatically delete their children. We just need to delete from top-level down.
func (s *Service) deleteInstanceChildren(ctx context.Context, tx *ent.Tx, instanceID string) error {
	// Delete direct children using Has...With predicates
	// The CASCADE DELETE on edges will automatically remove nested children

	// Disks (CASCADE will delete licenses)
	_, err := tx.BronzeGCPComputeInstanceDisk.Delete().
		Where(bronzegcpcomputeinstancedisk.HasInstanceWith(bronzegcpcomputeinstance.ID(instanceID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete disks: %w", err)
	}

	// NICs (CASCADE will delete access configs and alias ranges)
	_, err = tx.BronzeGCPComputeInstanceNIC.Delete().
		Where(bronzegcpcomputeinstancenic.HasInstanceWith(bronzegcpcomputeinstance.ID(instanceID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete NICs: %w", err)
	}

	// Labels
	_, err = tx.BronzeGCPComputeInstanceLabel.Delete().
		Where(bronzegcpcomputeinstancelabel.HasInstanceWith(bronzegcpcomputeinstance.ID(instanceID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete labels: %w", err)
	}

	// Tags
	_, err = tx.BronzeGCPComputeInstanceTag.Delete().
		Where(bronzegcpcomputeinstancetag.HasInstanceWith(bronzegcpcomputeinstance.ID(instanceID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete tags: %w", err)
	}

	// Metadata
	_, err = tx.BronzeGCPComputeInstanceMetadata.Delete().
		Where(bronzegcpcomputeinstancemetadata.HasInstanceWith(bronzegcpcomputeinstance.ID(instanceID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	// Service Accounts
	_, err = tx.BronzeGCPComputeInstanceServiceAccount.Delete().
		Where(bronzegcpcomputeinstanceserviceaccount.HasInstanceWith(bronzegcpcomputeinstance.ID(instanceID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete service accounts: %w", err)
	}

	return nil
}

// createInstanceChildren creates all child entities for an instance.
func (s *Service) createInstanceChildren(ctx context.Context, tx *ent.Tx, instance *ent.BronzeGCPComputeInstance, data *InstanceData) error {
	// Create disks and their licenses
	for _, diskData := range data.Disks {
		diskCreate := tx.BronzeGCPComputeInstanceDisk.Create().
			SetInstance(instance).
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
			diskCreate.SetDiskEncryptionKeyJSON(diskData.DiskEncryptionKeyJSON)
		}
		if diskData.InitializeParamsJSON != nil {
			diskCreate.SetInitializeParamsJSON(diskData.InitializeParamsJSON)
		}

		savedDisk, err := diskCreate.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create disk: %w", err)
		}

		// Create licenses for this disk
		for _, licData := range diskData.Licenses {
			_, err := tx.BronzeGCPComputeInstanceDiskLicense.Create().
				SetDisk(savedDisk).
				SetLicense(licData.License).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create disk license: %w", err)
			}
		}
	}

	// Create NICs and their nested children
	for _, nicData := range data.NICs {
		nicCreate := tx.BronzeGCPComputeInstanceNIC.Create().
			SetInstance(instance).
			SetName(nicData.Name).
			SetNetwork(nicData.Network).
			SetSubnetwork(nicData.Subnetwork).
			SetNetworkIP(nicData.NetworkIP).
			SetStackType(nicData.StackType).
			SetNicType(nicData.NicType)

		savedNIC, err := nicCreate.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create NIC: %w", err)
		}

		// Create access configs for this NIC
		for _, acData := range nicData.AccessConfigs {
			_, err := tx.BronzeGCPComputeInstanceNICAccessConfig.Create().
				SetNic(savedNIC).
				SetType(acData.Type).
				SetName(acData.Name).
				SetNatIP(acData.NatIP).
				SetNetworkTier(acData.NetworkTier).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create access config: %w", err)
			}
		}

		// Create alias ranges for this NIC
		for _, arData := range nicData.AliasIPRanges {
			_, err := tx.BronzeGCPComputeInstanceNICAliasRange.Create().
				SetNic(savedNIC).
				SetIPCidrRange(arData.IPCidrRange).
				SetSubnetworkRangeName(arData.SubnetworkRangeName).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create alias range: %w", err)
			}
		}
	}

	// Create labels
	for _, labelData := range data.Labels {
		_, err := tx.BronzeGCPComputeInstanceLabel.Create().
			SetInstance(instance).
			SetKey(labelData.Key).
			SetValue(labelData.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label: %w", err)
		}
	}

	// Create tags
	for _, tagData := range data.Tags {
		_, err := tx.BronzeGCPComputeInstanceTag.Create().
			SetInstance(instance).
			SetTag(tagData.Tag).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create tag: %w", err)
		}
	}

	// Create metadata
	for _, metaData := range data.Metadata {
		_, err := tx.BronzeGCPComputeInstanceMetadata.Create().
			SetInstance(instance).
			SetKey(metaData.Key).
			SetValue(metaData.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create metadata: %w", err)
		}
	}

	// Create service accounts
	for _, saData := range data.ServiceAccounts {
		saCreate := tx.BronzeGCPComputeInstanceServiceAccount.Create().
			SetInstance(instance).
			SetEmail(saData.Email)

		if saData.ScopesJSON != nil {
			saCreate.SetScopesJSON(saData.ScopesJSON)
		}

		_, err := saCreate.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create service account: %w", err)
		}
	}

	return nil
}

// DeleteStaleInstances removes instances that were not collected in the latest run.
// Also closes history records for deleted instances.
func (s *Service) DeleteStaleInstances(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	// Find stale instances
	staleInstances, err := tx.BronzeGCPComputeInstance.Query().
		Where(
			bronzegcpcomputeinstance.ProjectID(projectID),
			bronzegcpcomputeinstance.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale instance
	for _, inst := range staleInstances {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, inst.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for instance %s: %w", inst.ID, err)
		}

		// Delete children (CASCADE will handle nested children)
		if err := s.deleteInstanceChildren(ctx, tx, inst.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete children for instance %s: %w", inst.ID, err)
		}

		// Delete instance
		if err := tx.BronzeGCPComputeInstance.DeleteOne(inst).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete instance %s: %w", inst.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
