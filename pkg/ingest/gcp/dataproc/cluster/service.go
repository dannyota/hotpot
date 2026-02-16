package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpdataproccluster"
)

// Service handles Dataproc cluster ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new Dataproc cluster ingestion service.
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

// Ingest fetches Dataproc clusters from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	clusters, err := s.client.ListClusters(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list Dataproc clusters: %w", err)
	}

	clusterDataList := make([]*ClusterData, 0, len(clusters))
	for _, c := range clusters {
		data, err := ConvertCluster(c, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert Dataproc cluster: %w", err)
		}
		clusterDataList = append(clusterDataList, data)
	}

	if err := s.saveClusters(ctx, clusterDataList); err != nil {
		return nil, fmt.Errorf("failed to save Dataproc clusters: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		ClusterCount:   len(clusterDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveClusters saves Dataproc clusters to the database with history tracking.
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
		existing, err := tx.BronzeGCPDataprocCluster.Query().
			Where(bronzegcpdataproccluster.ID(clusterData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing Dataproc cluster %s: %w", clusterData.ID, err)
		}

		diff := DiffClusterData(existing, clusterData)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPDataprocCluster.UpdateOneID(clusterData.ID).
				SetCollectedAt(clusterData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for Dataproc cluster %s: %w", clusterData.ID, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeGCPDataprocCluster.Create().
				SetID(clusterData.ID).
				SetClusterName(clusterData.ClusterName).
				SetProjectID(clusterData.ProjectID).
				SetLocation(clusterData.Location).
				SetCollectedAt(clusterData.CollectedAt).
				SetFirstCollectedAt(clusterData.CollectedAt)

			if clusterData.ClusterUUID != "" {
				create.SetClusterUUID(clusterData.ClusterUUID)
			}
			if clusterData.ConfigJSON != nil {
				create.SetConfigJSON(clusterData.ConfigJSON)
			}
			if clusterData.StatusJSON != nil {
				create.SetStatusJSON(clusterData.StatusJSON)
			}
			if clusterData.StatusHistoryJSON != nil {
				create.SetStatusHistoryJSON(clusterData.StatusHistoryJSON)
			}
			if clusterData.LabelsJSON != nil {
				create.SetLabelsJSON(clusterData.LabelsJSON)
			}
			if clusterData.MetricsJSON != nil {
				create.SetMetricsJSON(clusterData.MetricsJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create Dataproc cluster %s: %w", clusterData.ID, err)
			}
		} else {
			update := tx.BronzeGCPDataprocCluster.UpdateOneID(clusterData.ID).
				SetClusterName(clusterData.ClusterName).
				SetProjectID(clusterData.ProjectID).
				SetLocation(clusterData.Location).
				SetCollectedAt(clusterData.CollectedAt)

			if clusterData.ClusterUUID != "" {
				update.SetClusterUUID(clusterData.ClusterUUID)
			}
			if clusterData.ConfigJSON != nil {
				update.SetConfigJSON(clusterData.ConfigJSON)
			}
			if clusterData.StatusJSON != nil {
				update.SetStatusJSON(clusterData.StatusJSON)
			}
			if clusterData.StatusHistoryJSON != nil {
				update.SetStatusHistoryJSON(clusterData.StatusHistoryJSON)
			}
			if clusterData.LabelsJSON != nil {
				update.SetLabelsJSON(clusterData.LabelsJSON)
			}
			if clusterData.MetricsJSON != nil {
				update.SetMetricsJSON(clusterData.MetricsJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update Dataproc cluster %s: %w", clusterData.ID, err)
			}
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, clusterData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for Dataproc cluster %s: %w", clusterData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, clusterData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for Dataproc cluster %s: %w", clusterData.ID, err)
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

	staleClusters, err := tx.BronzeGCPDataprocCluster.Query().
		Where(
			bronzegcpdataproccluster.ProjectID(projectID),
			bronzegcpdataproccluster.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, cluster := range staleClusters {
		if err := s.history.CloseHistory(ctx, tx, cluster.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for Dataproc cluster %s: %w", cluster.ID, err)
		}

		if err := tx.BronzeGCPDataprocCluster.DeleteOne(cluster).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete Dataproc cluster %s: %w", cluster.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
