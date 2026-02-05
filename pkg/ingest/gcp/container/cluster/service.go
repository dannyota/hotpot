package cluster

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
)

// Service handles GCP Container cluster ingestion.
type Service struct {
	client  *Client
	db      *gorm.DB
	history *HistoryService
}

// NewService creates a new cluster ingestion service.
func NewService(client *Client, db *gorm.DB) *Service {
	return &Service{
		client:  client,
		db:      db,
		history: NewHistoryService(db),
	}
}

// IngestParams contains parameters for cluster ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of cluster ingestion.
type IngestResult struct {
	ProjectID      string
	ClusterCount   int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches clusters from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch clusters from GCP
	clusters, err := s.client.ListClusters(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}

	// Convert to bronze models
	bronzeClusters := make([]bronze.GCPContainerCluster, 0, len(clusters))
	for _, cluster := range clusters {
		bronzeClusters = append(bronzeClusters, ConvertCluster(cluster, params.ProjectID, collectedAt))
	}

	// Save to database
	if err := s.saveClusters(ctx, bronzeClusters); err != nil {
		return nil, fmt.Errorf("failed to save clusters: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		ClusterCount:   len(bronzeClusters),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveClusters saves clusters to the database with history tracking.
func (s *Service) saveClusters(ctx context.Context, clusters []bronze.GCPContainerCluster) error {
	if len(clusters) == 0 {
		return nil
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, cluster := range clusters {
			// Load existing cluster with all relations
			var existing *bronze.GCPContainerCluster
			var old bronze.GCPContainerCluster
			err := tx.Preload("Labels").Preload("Addons").Preload("Conditions").Preload("NodePools").
				Where("resource_id = ?", cluster.ResourceID).
				First(&old).Error
			if err == nil {
				existing = &old
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to load existing cluster %s: %w", cluster.Name, err)
			}

			// Compute diff
			diff := DiffCluster(existing, &cluster)

			// Skip if no changes
			if !diff.HasAnyChange() && existing != nil {
				// Update collected_at only
				if err := tx.Model(&bronze.GCPContainerCluster{}).
					Where("resource_id = ?", cluster.ResourceID).
					Update("collected_at", cluster.CollectedAt).Error; err != nil {
					return fmt.Errorf("failed to update collected_at for cluster %s: %w", cluster.Name, err)
				}
				continue
			}

			// Delete old relations (manual cascade)
			if existing != nil {
				if err := s.deleteClusterRelations(tx, cluster.ResourceID); err != nil {
					return fmt.Errorf("failed to delete old relations for cluster %s: %w", cluster.Name, err)
				}
			}

			// Upsert cluster
			if err := tx.Save(&cluster).Error; err != nil {
				return fmt.Errorf("failed to upsert cluster %s: %w", cluster.Name, err)
			}

			// Create new relations
			if err := s.createClusterRelations(tx, cluster.ResourceID, &cluster); err != nil {
				return fmt.Errorf("failed to create relations for cluster %s: %w", cluster.Name, err)
			}

			// Track history
			if diff.IsNew {
				if err := s.history.CreateHistory(tx, &cluster, now); err != nil {
					return fmt.Errorf("failed to create history for cluster %s: %w", cluster.Name, err)
				}
			} else {
				if err := s.history.UpdateHistory(tx, existing, &cluster, diff, now); err != nil {
					return fmt.Errorf("failed to update history for cluster %s: %w", cluster.Name, err)
				}
			}
		}

		return nil
	})
}

// deleteClusterRelations deletes all related records for a cluster.
func (s *Service) deleteClusterRelations(tx *gorm.DB, clusterResourceID string) error {
	tables := []interface{}{
		&bronze.GCPContainerClusterLabel{},
		&bronze.GCPContainerClusterAddon{},
		&bronze.GCPContainerClusterCondition{},
		&bronze.GCPContainerClusterNodePool{},
	}

	for _, table := range tables {
		if err := tx.Where("cluster_resource_id = ?", clusterResourceID).Delete(table).Error; err != nil {
			return err
		}
	}

	return nil
}

// createClusterRelations creates all related records for a cluster.
func (s *Service) createClusterRelations(tx *gorm.DB, clusterResourceID string, cluster *bronze.GCPContainerCluster) error {
	// Create labels
	for i := range cluster.Labels {
		cluster.Labels[i].ClusterResourceID = clusterResourceID
	}
	if len(cluster.Labels) > 0 {
		if err := tx.Create(&cluster.Labels).Error; err != nil {
			return fmt.Errorf("failed to create labels: %w", err)
		}
	}

	// Create addons
	for i := range cluster.Addons {
		cluster.Addons[i].ClusterResourceID = clusterResourceID
	}
	if len(cluster.Addons) > 0 {
		if err := tx.Create(&cluster.Addons).Error; err != nil {
			return fmt.Errorf("failed to create addons: %w", err)
		}
	}

	// Create conditions
	for i := range cluster.Conditions {
		cluster.Conditions[i].ClusterResourceID = clusterResourceID
	}
	if len(cluster.Conditions) > 0 {
		if err := tx.Create(&cluster.Conditions).Error; err != nil {
			return fmt.Errorf("failed to create conditions: %w", err)
		}
	}

	// Create node pools
	for i := range cluster.NodePools {
		cluster.NodePools[i].ClusterResourceID = clusterResourceID
	}
	if len(cluster.NodePools) > 0 {
		if err := tx.Create(&cluster.NodePools).Error; err != nil {
			return fmt.Errorf("failed to create node pools: %w", err)
		}
	}

	return nil
}

// DeleteStaleClusters removes clusters that were not collected in the latest run.
// Also closes history records for deleted clusters.
func (s *Service) DeleteStaleClusters(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find stale clusters
		var staleClusters []bronze.GCPContainerCluster
		if err := tx.Where("project_id = ? AND collected_at < ?", projectID, collectedAt).
			Find(&staleClusters).Error; err != nil {
			return err
		}

		// Close history and delete each stale cluster
		for _, cluster := range staleClusters {
			// Close history
			if err := s.history.CloseHistory(tx, cluster.ResourceID, now); err != nil {
				return fmt.Errorf("failed to close history for cluster %s: %w", cluster.ResourceID, err)
			}

			// Delete relations
			if err := s.deleteClusterRelations(tx, cluster.ResourceID); err != nil {
				return fmt.Errorf("failed to delete relations for cluster %s: %w", cluster.ResourceID, err)
			}

			// Delete cluster
			if err := tx.Delete(&cluster).Error; err != nil {
				return fmt.Errorf("failed to delete cluster %s: %w", cluster.ResourceID, err)
			}
		}

		return nil
	})
}
