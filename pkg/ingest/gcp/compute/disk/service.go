package disk

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcomputedisk"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcomputedisklabel"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcomputedisklicense"
)

// Service handles GCP Compute disk ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new disk ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for disk ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of disk ingestion.
type IngestResult struct {
	ProjectID      string
	DiskCount      int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches disks from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch disks from GCP
	disks, err := s.client.ListDisks(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list disks: %w", err)
	}

	// Convert to data structs
	diskDataList := make([]*DiskData, 0, len(disks))
	for _, d := range disks {
		data, err := ConvertDisk(d, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert disk: %w", err)
		}
		diskDataList = append(diskDataList, data)
	}

	// Save to database
	if err := s.saveDisks(ctx, diskDataList); err != nil {
		return nil, fmt.Errorf("failed to save disks: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		DiskCount:      len(diskDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveDisks saves disks to the database with history tracking.
func (s *Service) saveDisks(ctx context.Context, disks []*DiskData) error {
	if len(disks) == 0 {
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

	for _, diskData := range disks {
		// Load existing disk with all edges
		existing, err := tx.BronzeGCPComputeDisk.Query().
			Where(bronzegcpcomputedisk.ID(diskData.ID)).
			WithLabels().
			WithLicenses().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing disk %s: %w", diskData.Name, err)
		}

		// Compute diff
		diff := DiffDiskData(existing, diskData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPComputeDisk.UpdateOneID(diskData.ID).
				SetCollectedAt(diskData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for disk %s: %w", diskData.Name, err)
			}
			continue
		}

		// Delete old child entities if updating
		if existing != nil {
			if err := s.deleteDiskChildren(ctx, tx, diskData.ID); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old children for disk %s: %w", diskData.Name, err)
			}
		}

		// Create or update disk
		var savedDisk *ent.BronzeGCPComputeDisk
		if existing == nil {
			// Create new disk
			create := tx.BronzeGCPComputeDisk.Create().
				SetID(diskData.ID).
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
				SetProjectID(diskData.ProjectID).
				SetCollectedAt(diskData.CollectedAt).
				SetFirstCollectedAt(diskData.CollectedAt)

			if diskData.DiskEncryptionKeyJSON != nil {
				create.SetDiskEncryptionKeyJSON(diskData.DiskEncryptionKeyJSON)
			}
			if diskData.UsersJSON != nil {
				create.SetUsersJSON(diskData.UsersJSON)
			}
			if diskData.ReplicaZonesJSON != nil {
				create.SetReplicaZonesJSON(diskData.ReplicaZonesJSON)
			}
			if diskData.ResourcePoliciesJSON != nil {
				create.SetResourcePoliciesJSON(diskData.ResourcePoliciesJSON)
			}
			if diskData.GuestOsFeaturesJSON != nil {
				create.SetGuestOsFeaturesJSON(diskData.GuestOsFeaturesJSON)
			}

			savedDisk, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create disk %s: %w", diskData.Name, err)
			}
		} else {
			// Update existing disk
			update := tx.BronzeGCPComputeDisk.UpdateOneID(diskData.ID).
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
				SetProjectID(diskData.ProjectID).
				SetCollectedAt(diskData.CollectedAt)

			if diskData.DiskEncryptionKeyJSON != nil {
				update.SetDiskEncryptionKeyJSON(diskData.DiskEncryptionKeyJSON)
			}
			if diskData.UsersJSON != nil {
				update.SetUsersJSON(diskData.UsersJSON)
			}
			if diskData.ReplicaZonesJSON != nil {
				update.SetReplicaZonesJSON(diskData.ReplicaZonesJSON)
			}
			if diskData.ResourcePoliciesJSON != nil {
				update.SetResourcePoliciesJSON(diskData.ResourcePoliciesJSON)
			}
			if diskData.GuestOsFeaturesJSON != nil {
				update.SetGuestOsFeaturesJSON(diskData.GuestOsFeaturesJSON)
			}

			savedDisk, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update disk %s: %w", diskData.Name, err)
			}
		}

		// Create child entities
		if err := s.createDiskChildren(ctx, tx, savedDisk, diskData); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create children for disk %s: %w", diskData.Name, err)
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, diskData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for disk %s: %w", diskData.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, diskData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for disk %s: %w", diskData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// deleteDiskChildren deletes all child entities for a disk.
func (s *Service) deleteDiskChildren(ctx context.Context, tx *ent.Tx, diskID string) error {
	// Delete labels
	_, err := tx.BronzeGCPComputeDiskLabel.Delete().
		Where(bronzegcpcomputedisklabel.HasDiskWith(bronzegcpcomputedisk.ID(diskID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete labels: %w", err)
	}

	// Delete licenses
	_, err = tx.BronzeGCPComputeDiskLicense.Delete().
		Where(bronzegcpcomputedisklicense.HasDiskWith(bronzegcpcomputedisk.ID(diskID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete licenses: %w", err)
	}

	return nil
}

// createDiskChildren creates all child entities for a disk.
func (s *Service) createDiskChildren(ctx context.Context, tx *ent.Tx, disk *ent.BronzeGCPComputeDisk, data *DiskData) error {
	// Create labels
	for _, labelData := range data.Labels {
		_, err := tx.BronzeGCPComputeDiskLabel.Create().
			SetDisk(disk).
			SetKey(labelData.Key).
			SetValue(labelData.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label %s: %w", labelData.Key, err)
		}
	}

	// Create licenses
	for _, licenseData := range data.Licenses {
		_, err := tx.BronzeGCPComputeDiskLicense.Create().
			SetDisk(disk).
			SetLicense(licenseData.License).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create license %s: %w", licenseData.License, err)
		}
	}

	return nil
}

// DeleteStaleDisks removes disks that were not collected in the latest run.
// Also closes history records for deleted disks.
func (s *Service) DeleteStaleDisks(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale disks
	staleDisks, err := tx.BronzeGCPComputeDisk.Query().
		Where(
			bronzegcpcomputedisk.ProjectID(projectID),
			bronzegcpcomputedisk.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to query stale disks: %w", err)
	}

	// Close history and delete each stale disk
	for _, d := range staleDisks {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, d.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for disk %s: %w", d.ID, err)
		}

		// Delete children
		if err := s.deleteDiskChildren(ctx, tx, d.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete children for disk %s: %w", d.ID, err)
		}

		// Delete disk
		if err := tx.BronzeGCPComputeDisk.DeleteOneID(d.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete disk %s: %w", d.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
