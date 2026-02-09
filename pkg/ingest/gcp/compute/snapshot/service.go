package snapshot

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcomputesnapshot"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcomputesnapshotlabel"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcomputesnapshotlicense"
)

// Service handles GCP Compute snapshot ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new snapshot ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for snapshot ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of snapshot ingestion.
type IngestResult struct {
	ProjectID      string
	SnapshotCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches snapshots from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch snapshots from GCP
	snapshots, err := s.client.ListSnapshots(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}

	// Convert to data structs
	snapshotDataList := make([]*SnapshotData, 0, len(snapshots))
	for _, snap := range snapshots {
		data, err := ConvertSnapshot(snap, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert snapshot: %w", err)
		}
		snapshotDataList = append(snapshotDataList, data)
	}

	// Save to database
	if err := s.saveSnapshots(ctx, snapshotDataList); err != nil {
		return nil, fmt.Errorf("failed to save snapshots: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		SnapshotCount:  len(snapshotDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveSnapshots saves snapshots to the database with history tracking.
func (s *Service) saveSnapshots(ctx context.Context, snapshots []*SnapshotData) error {
	if len(snapshots) == 0 {
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

	for _, snapshotData := range snapshots {
		// Load existing snapshot with all edges
		existing, err := tx.BronzeGCPComputeSnapshot.Query().
			Where(bronzegcpcomputesnapshot.ID(snapshotData.ID)).
			WithLabels().
			WithLicenses().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing snapshot %s: %w", snapshotData.Name, err)
		}

		// Compute diff
		diff := DiffSnapshotData(existing, snapshotData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPComputeSnapshot.UpdateOneID(snapshotData.ID).
				SetCollectedAt(snapshotData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for snapshot %s: %w", snapshotData.Name, err)
			}
			continue
		}

		// Delete old child entities if updating
		if existing != nil {
			if err := s.deleteSnapshotChildren(ctx, tx, snapshotData.ID); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old children for snapshot %s: %w", snapshotData.Name, err)
			}
		}

		// Create or update snapshot
		var savedSnapshot *ent.BronzeGCPComputeSnapshot
		if existing == nil {
			// Create new snapshot
			create := tx.BronzeGCPComputeSnapshot.Create().
				SetID(snapshotData.ID).
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
				SetProjectID(snapshotData.ProjectID).
				SetCollectedAt(snapshotData.CollectedAt).
				SetFirstCollectedAt(snapshotData.CollectedAt)

			if snapshotData.SnapshotEncryptionKeyJSON != nil {
				create.SetSnapshotEncryptionKeyJSON(snapshotData.SnapshotEncryptionKeyJSON)
			}
			if snapshotData.SourceDiskEncryptionKeyJSON != nil {
				create.SetSourceDiskEncryptionKeyJSON(snapshotData.SourceDiskEncryptionKeyJSON)
			}
			if snapshotData.GuestOsFeaturesJSON != nil {
				create.SetGuestOsFeaturesJSON(snapshotData.GuestOsFeaturesJSON)
			}
			if snapshotData.StorageLocationsJSON != nil {
				create.SetStorageLocationsJSON(snapshotData.StorageLocationsJSON)
			}

			savedSnapshot, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create snapshot %s: %w", snapshotData.Name, err)
			}
		} else {
			// Update existing snapshot
			update := tx.BronzeGCPComputeSnapshot.UpdateOneID(snapshotData.ID).
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
				SetProjectID(snapshotData.ProjectID).
				SetCollectedAt(snapshotData.CollectedAt)

			if snapshotData.SnapshotEncryptionKeyJSON != nil {
				update.SetSnapshotEncryptionKeyJSON(snapshotData.SnapshotEncryptionKeyJSON)
			}
			if snapshotData.SourceDiskEncryptionKeyJSON != nil {
				update.SetSourceDiskEncryptionKeyJSON(snapshotData.SourceDiskEncryptionKeyJSON)
			}
			if snapshotData.GuestOsFeaturesJSON != nil {
				update.SetGuestOsFeaturesJSON(snapshotData.GuestOsFeaturesJSON)
			}
			if snapshotData.StorageLocationsJSON != nil {
				update.SetStorageLocationsJSON(snapshotData.StorageLocationsJSON)
			}

			savedSnapshot, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update snapshot %s: %w", snapshotData.Name, err)
			}
		}

		// Create child entities
		if err := s.createSnapshotChildren(ctx, tx, savedSnapshot, snapshotData); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create children for snapshot %s: %w", snapshotData.Name, err)
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, snapshotData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for snapshot %s: %w", snapshotData.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, snapshotData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for snapshot %s: %w", snapshotData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// deleteSnapshotChildren deletes all child entities for a snapshot.
func (s *Service) deleteSnapshotChildren(ctx context.Context, tx *ent.Tx, snapshotID string) error {
	// Delete labels
	_, err := tx.BronzeGCPComputeSnapshotLabel.Delete().
		Where(bronzegcpcomputesnapshotlabel.HasSnapshotWith(bronzegcpcomputesnapshot.ID(snapshotID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete labels: %w", err)
	}

	// Delete licenses
	_, err = tx.BronzeGCPComputeSnapshotLicense.Delete().
		Where(bronzegcpcomputesnapshotlicense.HasSnapshotWith(bronzegcpcomputesnapshot.ID(snapshotID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete licenses: %w", err)
	}

	return nil
}

// createSnapshotChildren creates all child entities for a snapshot.
func (s *Service) createSnapshotChildren(ctx context.Context, tx *ent.Tx, snapshot *ent.BronzeGCPComputeSnapshot, data *SnapshotData) error {
	// Create labels
	for _, labelData := range data.Labels {
		_, err := tx.BronzeGCPComputeSnapshotLabel.Create().
			SetSnapshot(snapshot).
			SetKey(labelData.Key).
			SetValue(labelData.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label %s: %w", labelData.Key, err)
		}
	}

	// Create licenses
	for _, licenseData := range data.Licenses {
		_, err := tx.BronzeGCPComputeSnapshotLicense.Create().
			SetSnapshot(snapshot).
			SetLicense(licenseData.License).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create license %s: %w", licenseData.License, err)
		}
	}

	return nil
}

// DeleteStaleSnapshots removes snapshots that were not collected in the latest run.
// Also closes history records for deleted snapshots.
func (s *Service) DeleteStaleSnapshots(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale snapshots
	staleSnapshots, err := tx.BronzeGCPComputeSnapshot.Query().
		Where(
			bronzegcpcomputesnapshot.ProjectID(projectID),
			bronzegcpcomputesnapshot.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to query stale snapshots: %w", err)
	}

	// Close history and delete each stale snapshot
	for _, snap := range staleSnapshots {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, snap.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for snapshot %s: %w", snap.ID, err)
		}

		// Delete children
		if err := s.deleteSnapshotChildren(ctx, tx, snap.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete children for snapshot %s: %w", snap.ID, err)
		}

		// Delete snapshot
		if err := tx.BronzeGCPComputeSnapshot.DeleteOneID(snap.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete snapshot %s: %w", snap.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
