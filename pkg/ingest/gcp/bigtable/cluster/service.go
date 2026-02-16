package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpbigtablecluster"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpbigtableinstance"
)

// Service handles Bigtable cluster ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new Bigtable cluster ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
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

// Ingest fetches Bigtable clusters from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch clusters from GCP (queries instances from DB, then fetches clusters per instance)
	rawClusters, err := s.client.ListClusters(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}

	// Convert to data structs
	clusterDataList := make([]*ClusterData, 0, len(rawClusters))
	for _, raw := range rawClusters {
		data, err := ConvertCluster(raw, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert cluster: %w", err)
		}
		clusterDataList = append(clusterDataList, data)
	}

	// Save to database
	if err := s.saveClusters(ctx, clusterDataList); err != nil {
		return nil, fmt.Errorf("failed to save clusters: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		ClusterCount:   len(clusterDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveClusters saves Bigtable clusters to the database with history tracking.
func (s *Service) saveClusters(ctx context.Context, clusters []*ClusterData) error {
	if len(clusters) == 0 {
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

	for _, clusterData := range clusters {
		// Look up the parent instance for edge creation
		parentInstance, err := tx.BronzeGCPBigtableInstance.Query().
			Where(bronzegcpbigtableinstance.ID(clusterData.InstanceName)).
			First(ctx)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to find parent instance %s: %w", clusterData.InstanceName, err)
		}

		// Load existing cluster
		existing, err := tx.BronzeGCPBigtableCluster.Query().
			Where(bronzegcpbigtablecluster.ID(clusterData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing cluster %s: %w", clusterData.ID, err)
		}

		// Compute diff
		diff := DiffClusterData(existing, clusterData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPBigtableCluster.UpdateOneID(clusterData.ID).
				SetCollectedAt(clusterData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for cluster %s: %w", clusterData.ID, err)
			}
			continue
		}

		// Create or update cluster
		if existing == nil {
			create := tx.BronzeGCPBigtableCluster.Create().
				SetID(clusterData.ID).
				SetLocation(clusterData.Location).
				SetState(clusterData.State).
				SetServeNodes(clusterData.ServeNodes).
				SetDefaultStorageType(clusterData.DefaultStorageType).
				SetInstanceName(clusterData.InstanceName).
				SetProjectID(clusterData.ProjectID).
				SetCollectedAt(clusterData.CollectedAt).
				SetFirstCollectedAt(clusterData.CollectedAt).
				SetInstance(parentInstance)

			if clusterData.EncryptionConfigJSON != nil {
				create.SetEncryptionConfigJSON(clusterData.EncryptionConfigJSON)
			}
			if clusterData.ClusterConfigJSON != nil {
				create.SetClusterConfigJSON(clusterData.ClusterConfigJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create cluster %s: %w", clusterData.ID, err)
			}
		} else {
			update := tx.BronzeGCPBigtableCluster.UpdateOneID(clusterData.ID).
				SetLocation(clusterData.Location).
				SetState(clusterData.State).
				SetServeNodes(clusterData.ServeNodes).
				SetDefaultStorageType(clusterData.DefaultStorageType).
				SetInstanceName(clusterData.InstanceName).
				SetProjectID(clusterData.ProjectID).
				SetCollectedAt(clusterData.CollectedAt).
				SetInstance(parentInstance)

			if clusterData.EncryptionConfigJSON != nil {
				update.SetEncryptionConfigJSON(clusterData.EncryptionConfigJSON)
			}
			if clusterData.ClusterConfigJSON != nil {
				update.SetClusterConfigJSON(clusterData.ClusterConfigJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update cluster %s: %w", clusterData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, clusterData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for cluster %s: %w", clusterData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, clusterData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for cluster %s: %w", clusterData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleClusters removes clusters that were not collected in the latest run.
func (s *Service) DeleteStaleClusters(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	staleClusters, err := tx.BronzeGCPBigtableCluster.Query().
		Where(
			bronzegcpbigtablecluster.ProjectID(projectID),
			bronzegcpbigtablecluster.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, c := range staleClusters {
		if err := s.history.CloseHistory(ctx, tx, c.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for cluster %s: %w", c.ID, err)
		}

		if err := tx.BronzeGCPBigtableCluster.DeleteOne(c).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete cluster %s: %w", c.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
