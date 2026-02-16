package instance

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpsqlinstance"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpsqlinstancelabel"
)

// Service handles GCP Cloud SQL instance ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new SQL instance ingestion service.
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

// Ingest fetches SQL instances from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch instances from GCP
	instances, err := s.client.ListInstances(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list SQL instances: %w", err)
	}

	// Convert to data structs
	instanceDataList := make([]*InstanceData, 0, len(instances))
	for _, inst := range instances {
		data, err := ConvertInstance(inst, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert SQL instance: %w", err)
		}
		instanceDataList = append(instanceDataList, data)
	}

	// Save to database
	if err := s.saveInstances(ctx, instanceDataList); err != nil {
		return nil, fmt.Errorf("failed to save SQL instances: %w", err)
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
		// Load existing instance with labels
		existing, err := tx.BronzeGCPSQLInstance.Query().
			Where(bronzegcpsqlinstance.ID(instanceData.ResourceID)).
			WithLabels().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing SQL instance %s: %w", instanceData.Name, err)
		}

		// Compute diff
		diff := DiffInstanceData(existing, instanceData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPSQLInstance.UpdateOneID(instanceData.ResourceID).
				SetCollectedAt(instanceData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for SQL instance %s: %w", instanceData.Name, err)
			}
			continue
		}

		// Delete old children if updating
		if existing != nil {
			if err := deleteInstanceChildren(ctx, tx, instanceData.ResourceID); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old children for SQL instance %s: %w", instanceData.Name, err)
			}
		}

		// Create or update instance
		var savedInstance *ent.BronzeGCPSQLInstance
		if existing == nil {
			// Create new instance
			create := tx.BronzeGCPSQLInstance.Create().
				SetID(instanceData.ResourceID).
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
				SetProjectID(instanceData.ProjectID).
				SetCollectedAt(instanceData.CollectedAt).
				SetFirstCollectedAt(instanceData.CollectedAt)

			if instanceData.SettingsJSON != nil {
				create.SetSettingsJSON(instanceData.SettingsJSON)
			}
			if instanceData.ServerCaCertJSON != nil {
				create.SetServerCaCertJSON(instanceData.ServerCaCertJSON)
			}
			if instanceData.IpAddressesJSON != nil {
				create.SetIPAddressesJSON(instanceData.IpAddressesJSON)
			}
			if instanceData.ReplicaConfigurationJSON != nil {
				create.SetReplicaConfigurationJSON(instanceData.ReplicaConfigurationJSON)
			}
			if instanceData.FailoverReplicaJSON != nil {
				create.SetFailoverReplicaJSON(instanceData.FailoverReplicaJSON)
			}
			if instanceData.DiskEncryptionConfigurationJSON != nil {
				create.SetDiskEncryptionConfigurationJSON(instanceData.DiskEncryptionConfigurationJSON)
			}
			if instanceData.DiskEncryptionStatusJSON != nil {
				create.SetDiskEncryptionStatusJSON(instanceData.DiskEncryptionStatusJSON)
			}

			savedInstance, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create SQL instance %s: %w", instanceData.Name, err)
			}
		} else {
			// Update existing instance
			update := tx.BronzeGCPSQLInstance.UpdateOneID(instanceData.ResourceID).
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
				SetProjectID(instanceData.ProjectID).
				SetCollectedAt(instanceData.CollectedAt)

			if instanceData.SettingsJSON != nil {
				update.SetSettingsJSON(instanceData.SettingsJSON)
			}
			if instanceData.ServerCaCertJSON != nil {
				update.SetServerCaCertJSON(instanceData.ServerCaCertJSON)
			}
			if instanceData.IpAddressesJSON != nil {
				update.SetIPAddressesJSON(instanceData.IpAddressesJSON)
			}
			if instanceData.ReplicaConfigurationJSON != nil {
				update.SetReplicaConfigurationJSON(instanceData.ReplicaConfigurationJSON)
			}
			if instanceData.FailoverReplicaJSON != nil {
				update.SetFailoverReplicaJSON(instanceData.FailoverReplicaJSON)
			}
			if instanceData.DiskEncryptionConfigurationJSON != nil {
				update.SetDiskEncryptionConfigurationJSON(instanceData.DiskEncryptionConfigurationJSON)
			}
			if instanceData.DiskEncryptionStatusJSON != nil {
				update.SetDiskEncryptionStatusJSON(instanceData.DiskEncryptionStatusJSON)
			}

			savedInstance, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update SQL instance %s: %w", instanceData.Name, err)
			}
		}

		// Create child entities
		if err := createInstanceChildren(ctx, tx, savedInstance, instanceData); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create children for SQL instance %s: %w", instanceData.Name, err)
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, instanceData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for SQL instance %s: %w", instanceData.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, instanceData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for SQL instance %s: %w", instanceData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// deleteInstanceChildren deletes all child entities for an instance.
func deleteInstanceChildren(ctx context.Context, tx *ent.Tx, instanceID string) error {
	// Delete labels
	_, err := tx.BronzeGCPSQLInstanceLabel.Delete().
		Where(bronzegcpsqlinstancelabel.HasInstanceWith(bronzegcpsqlinstance.ID(instanceID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete labels: %w", err)
	}

	return nil
}

// createInstanceChildren creates all child entities for an instance.
func createInstanceChildren(ctx context.Context, tx *ent.Tx, instance *ent.BronzeGCPSQLInstance, data *InstanceData) error {
	// Create labels
	for _, labelData := range data.Labels {
		_, err := tx.BronzeGCPSQLInstanceLabel.Create().
			SetInstance(instance).
			SetKey(labelData.Key).
			SetValue(labelData.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label: %w", err)
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
	staleInstances, err := tx.BronzeGCPSQLInstance.Query().
		Where(
			bronzegcpsqlinstance.ProjectID(projectID),
			bronzegcpsqlinstance.CollectedAtLT(collectedAt),
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
			return fmt.Errorf("failed to close history for SQL instance %s: %w", inst.ID, err)
		}

		// Delete instance (CASCADE will handle labels automatically)
		if err := tx.BronzeGCPSQLInstance.DeleteOne(inst).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete SQL instance %s: %w", inst.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
