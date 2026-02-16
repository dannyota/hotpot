package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpalloydbcluster"
)

// Service handles AlloyDB cluster ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new AlloyDB cluster ingestion service.
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

// Ingest fetches AlloyDB clusters from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	clusters, err := s.client.ListClusters(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}

	clusterDataList := make([]*ClusterData, 0, len(clusters))
	for _, c := range clusters {
		data, err := ConvertCluster(c, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert cluster: %w", err)
		}
		clusterDataList = append(clusterDataList, data)
	}

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
		existing, err := tx.BronzeGCPAlloyDBCluster.Query().
			Where(bronzegcpalloydbcluster.ID(clusterData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing cluster %s: %w", clusterData.ID, err)
		}

		diff := DiffClusterData(existing, clusterData)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPAlloyDBCluster.UpdateOneID(clusterData.ID).
				SetCollectedAt(clusterData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for cluster %s: %w", clusterData.ID, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeGCPAlloyDBCluster.Create().
				SetID(clusterData.ID).
				SetName(clusterData.Name).
				SetState(clusterData.State).
				SetClusterType(clusterData.ClusterType).
				SetDatabaseVersion(clusterData.DatabaseVersion).
				SetReconciling(clusterData.Reconciling).
				SetSatisfiesPzs(clusterData.SatisfiesPzs).
				SetSubscriptionType(clusterData.SubscriptionType).
				SetProjectID(clusterData.ProjectID).
				SetLocation(clusterData.Location).
				SetCollectedAt(clusterData.CollectedAt).
				SetFirstCollectedAt(clusterData.CollectedAt)

			if clusterData.DisplayName != "" {
				create.SetDisplayName(clusterData.DisplayName)
			}
			if clusterData.UID != "" {
				create.SetUID(clusterData.UID)
			}
			if clusterData.CreateTime != "" {
				create.SetCreateTime(clusterData.CreateTime)
			}
			if clusterData.UpdateTime != "" {
				create.SetUpdateTime(clusterData.UpdateTime)
			}
			if clusterData.DeleteTime != "" {
				create.SetDeleteTime(clusterData.DeleteTime)
			}
			if clusterData.Network != "" {
				create.SetNetwork(clusterData.Network)
			}
			if clusterData.Etag != "" {
				create.SetEtag(clusterData.Etag)
			}
			if clusterData.LabelsJSON != nil {
				create.SetLabelsJSON(clusterData.LabelsJSON)
			}
			if clusterData.NetworkConfigJSON != nil {
				create.SetNetworkConfigJSON(clusterData.NetworkConfigJSON)
			}
			if clusterData.AnnotationsJSON != nil {
				create.SetAnnotationsJSON(clusterData.AnnotationsJSON)
			}
			if clusterData.InitialUserJSON != nil {
				create.SetInitialUserJSON(clusterData.InitialUserJSON)
			}
			if clusterData.AutomatedBackupPolicyJSON != nil {
				create.SetAutomatedBackupPolicyJSON(clusterData.AutomatedBackupPolicyJSON)
			}
			if clusterData.SslConfigJSON != nil {
				create.SetSslConfigJSON(clusterData.SslConfigJSON)
			}
			if clusterData.EncryptionConfigJSON != nil {
				create.SetEncryptionConfigJSON(clusterData.EncryptionConfigJSON)
			}
			if clusterData.EncryptionInfoJSON != nil {
				create.SetEncryptionInfoJSON(clusterData.EncryptionInfoJSON)
			}
			if clusterData.ContinuousBackupConfigJSON != nil {
				create.SetContinuousBackupConfigJSON(clusterData.ContinuousBackupConfigJSON)
			}
			if clusterData.ContinuousBackupInfoJSON != nil {
				create.SetContinuousBackupInfoJSON(clusterData.ContinuousBackupInfoJSON)
			}
			if clusterData.SecondaryConfigJSON != nil {
				create.SetSecondaryConfigJSON(clusterData.SecondaryConfigJSON)
			}
			if clusterData.PrimaryConfigJSON != nil {
				create.SetPrimaryConfigJSON(clusterData.PrimaryConfigJSON)
			}
			if clusterData.PscConfigJSON != nil {
				create.SetPscConfigJSON(clusterData.PscConfigJSON)
			}
			if clusterData.MaintenanceUpdatePolicyJSON != nil {
				create.SetMaintenanceUpdatePolicyJSON(clusterData.MaintenanceUpdatePolicyJSON)
			}
			if clusterData.MaintenanceScheduleJSON != nil {
				create.SetMaintenanceScheduleJSON(clusterData.MaintenanceScheduleJSON)
			}
			if clusterData.TrialMetadataJSON != nil {
				create.SetTrialMetadataJSON(clusterData.TrialMetadataJSON)
			}
			if clusterData.TagsJSON != nil {
				create.SetTagsJSON(clusterData.TagsJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create cluster %s: %w", clusterData.ID, err)
			}
		} else {
			update := tx.BronzeGCPAlloyDBCluster.UpdateOneID(clusterData.ID).
				SetName(clusterData.Name).
				SetState(clusterData.State).
				SetClusterType(clusterData.ClusterType).
				SetDatabaseVersion(clusterData.DatabaseVersion).
				SetReconciling(clusterData.Reconciling).
				SetSatisfiesPzs(clusterData.SatisfiesPzs).
				SetSubscriptionType(clusterData.SubscriptionType).
				SetProjectID(clusterData.ProjectID).
				SetLocation(clusterData.Location).
				SetCollectedAt(clusterData.CollectedAt)

			if clusterData.DisplayName != "" {
				update.SetDisplayName(clusterData.DisplayName)
			}
			if clusterData.UID != "" {
				update.SetUID(clusterData.UID)
			}
			if clusterData.CreateTime != "" {
				update.SetCreateTime(clusterData.CreateTime)
			}
			if clusterData.UpdateTime != "" {
				update.SetUpdateTime(clusterData.UpdateTime)
			}
			if clusterData.DeleteTime != "" {
				update.SetDeleteTime(clusterData.DeleteTime)
			}
			if clusterData.Network != "" {
				update.SetNetwork(clusterData.Network)
			}
			if clusterData.Etag != "" {
				update.SetEtag(clusterData.Etag)
			}
			if clusterData.LabelsJSON != nil {
				update.SetLabelsJSON(clusterData.LabelsJSON)
			}
			if clusterData.NetworkConfigJSON != nil {
				update.SetNetworkConfigJSON(clusterData.NetworkConfigJSON)
			}
			if clusterData.AnnotationsJSON != nil {
				update.SetAnnotationsJSON(clusterData.AnnotationsJSON)
			}
			if clusterData.InitialUserJSON != nil {
				update.SetInitialUserJSON(clusterData.InitialUserJSON)
			}
			if clusterData.AutomatedBackupPolicyJSON != nil {
				update.SetAutomatedBackupPolicyJSON(clusterData.AutomatedBackupPolicyJSON)
			}
			if clusterData.SslConfigJSON != nil {
				update.SetSslConfigJSON(clusterData.SslConfigJSON)
			}
			if clusterData.EncryptionConfigJSON != nil {
				update.SetEncryptionConfigJSON(clusterData.EncryptionConfigJSON)
			}
			if clusterData.EncryptionInfoJSON != nil {
				update.SetEncryptionInfoJSON(clusterData.EncryptionInfoJSON)
			}
			if clusterData.ContinuousBackupConfigJSON != nil {
				update.SetContinuousBackupConfigJSON(clusterData.ContinuousBackupConfigJSON)
			}
			if clusterData.ContinuousBackupInfoJSON != nil {
				update.SetContinuousBackupInfoJSON(clusterData.ContinuousBackupInfoJSON)
			}
			if clusterData.SecondaryConfigJSON != nil {
				update.SetSecondaryConfigJSON(clusterData.SecondaryConfigJSON)
			}
			if clusterData.PrimaryConfigJSON != nil {
				update.SetPrimaryConfigJSON(clusterData.PrimaryConfigJSON)
			}
			if clusterData.PscConfigJSON != nil {
				update.SetPscConfigJSON(clusterData.PscConfigJSON)
			}
			if clusterData.MaintenanceUpdatePolicyJSON != nil {
				update.SetMaintenanceUpdatePolicyJSON(clusterData.MaintenanceUpdatePolicyJSON)
			}
			if clusterData.MaintenanceScheduleJSON != nil {
				update.SetMaintenanceScheduleJSON(clusterData.MaintenanceScheduleJSON)
			}
			if clusterData.TrialMetadataJSON != nil {
				update.SetTrialMetadataJSON(clusterData.TrialMetadataJSON)
			}
			if clusterData.TagsJSON != nil {
				update.SetTagsJSON(clusterData.TagsJSON)
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

	staleClusters, err := tx.BronzeGCPAlloyDBCluster.Query().
		Where(
			bronzegcpalloydbcluster.ProjectID(projectID),
			bronzegcpalloydbcluster.CollectedAtLT(collectedAt),
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

		if err := tx.BronzeGCPAlloyDBCluster.DeleteOne(c).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete cluster %s: %w", c.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
