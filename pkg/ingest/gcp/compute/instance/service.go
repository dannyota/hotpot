package instance

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
)

// Service handles GCP Compute instance ingestion.
type Service struct {
	client  *Client
	db      *gorm.DB
	history *HistoryService
}

// NewService creates a new instance ingestion service.
func NewService(client *Client, db *gorm.DB) *Service {
	return &Service{
		client:  client,
		db:      db,
		history: NewHistoryService(db),
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

	// Convert to bronze models
	bronzeInstances := make([]bronze.GCPComputeInstance, 0, len(instances))
	for _, inst := range instances {
		bronzeInstances = append(bronzeInstances, ConvertInstance(inst, params.ProjectID, collectedAt))
	}

	// Save to database
	if err := s.saveInstances(ctx, bronzeInstances); err != nil {
		return nil, fmt.Errorf("failed to save instances: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		InstanceCount:  len(bronzeInstances),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveInstances saves instances to the database with history tracking.
func (s *Service) saveInstances(ctx context.Context, instances []bronze.GCPComputeInstance) error {
	if len(instances) == 0 {
		return nil
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, instance := range instances {
			// Load existing instance with all relations
			var existing *bronze.GCPComputeInstance
			var old bronze.GCPComputeInstance
			err := tx.Preload("Disks").Preload("Disks.Licenses").
				Preload("NICs").Preload("NICs.AccessConfigs").Preload("NICs.AliasIpRanges").
				Preload("Labels").Preload("Tags").Preload("Metadata").Preload("ServiceAccounts").
				Where("resource_id = ?", instance.ResourceID).
				First(&old).Error
			if err == nil {
				existing = &old
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to load existing instance %s: %w", instance.Name, err)
			}

			// Compute diff
			diff := DiffInstance(existing, &instance)

			// Skip if no changes
			if !diff.HasAnyChange() && existing != nil {
				// Update collected_at only
				if err := tx.Model(&bronze.GCPComputeInstance{}).
					Where("resource_id = ?", instance.ResourceID).
					Update("collected_at", instance.CollectedAt).Error; err != nil {
					return fmt.Errorf("failed to update collected_at for instance %s: %w", instance.Name, err)
				}
				continue
			}

			// Delete old relations (manual cascade)
			if existing != nil {
				if err := s.deleteInstanceRelations(tx, instance.ResourceID); err != nil {
					return fmt.Errorf("failed to delete old relations for instance %s: %w", instance.Name, err)
				}
			}

			// Upsert instance
			if err := tx.Save(&instance).Error; err != nil {
				return fmt.Errorf("failed to upsert instance %s: %w", instance.Name, err)
			}

			// Create new relations
			if err := s.createInstanceRelations(tx, instance.ResourceID, &instance); err != nil {
				return fmt.Errorf("failed to create relations for instance %s: %w", instance.Name, err)
			}

			// Track history
			if diff.IsNew {
				if err := s.history.CreateHistory(tx, &instance, now); err != nil {
					return fmt.Errorf("failed to create history for instance %s: %w", instance.Name, err)
				}
			} else {
				if err := s.history.UpdateHistory(tx, existing, &instance, diff, now); err != nil {
					return fmt.Errorf("failed to update history for instance %s: %w", instance.Name, err)
				}
			}
		}

		return nil
	})
}

// deleteInstanceRelations deletes all related records for an instance.
func (s *Service) deleteInstanceRelations(tx *gorm.DB, instanceResourceID string) error {
	// Delete instance-level relations (linked by instance_resource_id)
	tables := []interface{}{
		&bronze.GCPComputeInstanceDisk{},
		&bronze.GCPComputeInstanceNIC{},
		&bronze.GCPComputeInstanceLabel{},
		&bronze.GCPComputeInstanceTag{},
		&bronze.GCPComputeInstanceMetadata{},
		&bronze.GCPComputeInstanceServiceAccount{},
	}

	for _, table := range tables {
		if err := tx.Where("instance_resource_id = ?", instanceResourceID).Delete(table).Error; err != nil {
			return err
		}
	}

	return nil
}

// createInstanceRelations creates all related records for an instance.
func (s *Service) createInstanceRelations(tx *gorm.DB, instanceResourceID string, instance *bronze.GCPComputeInstance) error {
	// Create disks and their nested relations
	for i := range instance.Disks {
		instance.Disks[i].InstanceResourceID = instanceResourceID
	}
	if len(instance.Disks) > 0 {
		if err := tx.Create(&instance.Disks).Error; err != nil {
			return fmt.Errorf("failed to create disks: %w", err)
		}

		// Create disk licenses (nested under disks, still use surrogate DiskID)
		for _, disk := range instance.Disks {
			for j := range disk.Licenses {
				disk.Licenses[j].DiskID = disk.ID
			}
			if len(disk.Licenses) > 0 {
				if err := tx.Create(&disk.Licenses).Error; err != nil {
					return fmt.Errorf("failed to create disk licenses: %w", err)
				}
			}
		}
	}

	// Create NICs and their nested relations
	for i := range instance.NICs {
		instance.NICs[i].InstanceResourceID = instanceResourceID
	}
	if len(instance.NICs) > 0 {
		if err := tx.Create(&instance.NICs).Error; err != nil {
			return fmt.Errorf("failed to create NICs: %w", err)
		}

		// Create NIC access configs and alias ranges (nested under NICs, still use surrogate NICID)
		for _, nic := range instance.NICs {
			for j := range nic.AccessConfigs {
				nic.AccessConfigs[j].NICID = nic.ID
			}
			if len(nic.AccessConfigs) > 0 {
				if err := tx.Create(&nic.AccessConfigs).Error; err != nil {
					return fmt.Errorf("failed to create NIC access configs: %w", err)
				}
			}

			for j := range nic.AliasIpRanges {
				nic.AliasIpRanges[j].NICID = nic.ID
			}
			if len(nic.AliasIpRanges) > 0 {
				if err := tx.Create(&nic.AliasIpRanges).Error; err != nil {
					return fmt.Errorf("failed to create NIC alias ranges: %w", err)
				}
			}
		}
	}

	// Create labels
	for i := range instance.Labels {
		instance.Labels[i].InstanceResourceID = instanceResourceID
	}
	if len(instance.Labels) > 0 {
		if err := tx.Create(&instance.Labels).Error; err != nil {
			return fmt.Errorf("failed to create labels: %w", err)
		}
	}

	// Create tags
	for i := range instance.Tags {
		instance.Tags[i].InstanceResourceID = instanceResourceID
	}
	if len(instance.Tags) > 0 {
		if err := tx.Create(&instance.Tags).Error; err != nil {
			return fmt.Errorf("failed to create tags: %w", err)
		}
	}

	// Create metadata
	for i := range instance.Metadata {
		instance.Metadata[i].InstanceResourceID = instanceResourceID
	}
	if len(instance.Metadata) > 0 {
		if err := tx.Create(&instance.Metadata).Error; err != nil {
			return fmt.Errorf("failed to create metadata: %w", err)
		}
	}

	// Create service accounts
	for i := range instance.ServiceAccounts {
		instance.ServiceAccounts[i].InstanceResourceID = instanceResourceID
	}
	if len(instance.ServiceAccounts) > 0 {
		if err := tx.Create(&instance.ServiceAccounts).Error; err != nil {
			return fmt.Errorf("failed to create service accounts: %w", err)
		}
	}

	return nil
}

// DeleteStaleInstances removes instances that were not collected in the latest run.
// Also closes history records for deleted instances.
func (s *Service) DeleteStaleInstances(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find stale instances
		var staleInstances []bronze.GCPComputeInstance
		if err := tx.Where("project_id = ? AND collected_at < ?", projectID, collectedAt).
			Find(&staleInstances).Error; err != nil {
			return err
		}

		// Close history and delete each stale instance
		for _, inst := range staleInstances {
			// Close history
			if err := s.history.CloseHistory(tx, inst.ResourceID, now); err != nil {
				return fmt.Errorf("failed to close history for instance %s: %w", inst.ResourceID, err)
			}

			// Delete relations
			if err := s.deleteInstanceRelations(tx, inst.ResourceID); err != nil {
				return fmt.Errorf("failed to delete relations for instance %s: %w", inst.ResourceID, err)
			}

			// Delete instance
			if err := tx.Delete(&inst).Error; err != nil {
				return fmt.Errorf("failed to delete instance %s: %w", inst.ResourceID, err)
			}
		}

		return nil
	})
}
