package cluster

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzegcpcontainercluster"
	"hotpot/pkg/storage/ent/bronzegcpcontainerclusteraddon"
	"hotpot/pkg/storage/ent/bronzegcpcontainerclustercondition"
	"hotpot/pkg/storage/ent/bronzegcpcontainerclusterlabel"
	"hotpot/pkg/storage/ent/bronzegcpcontainerclusternodepool"
)

// Service handles GCP Container cluster ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new cluster ingestion service.
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

// Ingest fetches clusters from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch clusters from GCP
	clusters, err := s.client.ListClusters(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}

	// Convert to data structs
	clusterDataList := make([]*ClusterData, 0, len(clusters))
	for _, cluster := range clusters {
		data, err := ConvertCluster(cluster, params.ProjectID, collectedAt)
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

// saveClusters saves clusters to the database with history tracking.
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
		// Load existing cluster with all nested edges
		existing, err := tx.BronzeGCPContainerCluster.Query().
			Where(bronzegcpcontainercluster.ID(clusterData.ResourceID)).
			WithLabels().
			WithAddons().
			WithConditions().
			WithNodePools().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing cluster %s: %w", clusterData.Name, err)
		}

		// Compute diff
		diff := DiffClusterData(existing, clusterData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPContainerCluster.UpdateOneID(clusterData.ResourceID).
				SetCollectedAt(clusterData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for cluster %s: %w", clusterData.Name, err)
			}
			continue
		}

		// Delete old child entities if updating
		if existing != nil {
			if err := s.deleteClusterChildren(ctx, tx, clusterData.ResourceID); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old children for cluster %s: %w", clusterData.Name, err)
			}
		}

		// Create or update cluster
		var savedCluster *ent.BronzeGCPContainerCluster
		if existing == nil {
			// Create new cluster
			create := tx.BronzeGCPContainerCluster.Create().
				SetID(clusterData.ResourceID).
				SetName(clusterData.Name).
				SetLocation(clusterData.Location).
				SetZone(clusterData.Zone).
				SetDescription(clusterData.Description).
				SetInitialClusterVersion(clusterData.InitialClusterVersion).
				SetCurrentMasterVersion(clusterData.CurrentMasterVersion).
				SetCurrentNodeVersion(clusterData.CurrentNodeVersion).
				SetStatus(clusterData.Status).
				SetStatusMessage(clusterData.StatusMessage).
				SetCurrentNodeCount(clusterData.CurrentNodeCount).
				SetNetwork(clusterData.Network).
				SetSubnetwork(clusterData.Subnetwork).
				SetClusterIpv4Cidr(clusterData.ClusterIpv4Cidr).
				SetServicesIpv4Cidr(clusterData.ServicesIpv4Cidr).
				SetNodeIpv4CidrSize(clusterData.NodeIpv4CidrSize).
				SetEndpoint(clusterData.Endpoint).
				SetSelfLink(clusterData.SelfLink).
				SetCreateTime(clusterData.CreateTime).
				SetExpireTime(clusterData.ExpireTime).
				SetEtag(clusterData.Etag).
				SetLabelFingerprint(clusterData.LabelFingerprint).
				SetLoggingService(clusterData.LoggingService).
				SetMonitoringService(clusterData.MonitoringService).
				SetEnableKubernetesAlpha(clusterData.EnableKubernetesAlpha).
				SetEnableTpu(clusterData.EnableTpu).
				SetTpuIpv4CidrBlock(clusterData.TpuIpv4CidrBlock).
				SetProjectID(clusterData.ProjectID).
				SetCollectedAt(clusterData.CollectedAt).
				SetFirstCollectedAt(clusterData.CollectedAt)

			// Set optional JSON fields
			if clusterData.AddonsConfigJSON != nil {
				create.SetAddonsConfigJSON(clusterData.AddonsConfigJSON)
			}
			if clusterData.PrivateClusterConfigJSON != nil {
				create.SetPrivateClusterConfigJSON(clusterData.PrivateClusterConfigJSON)
			}
			if clusterData.IPAllocationPolicyJSON != nil {
				create.SetIPAllocationPolicyJSON(clusterData.IPAllocationPolicyJSON)
			}
			if clusterData.NetworkConfigJSON != nil {
				create.SetNetworkConfigJSON(clusterData.NetworkConfigJSON)
			}
			if clusterData.MasterAuthJSON != nil {
				create.SetMasterAuthJSON(clusterData.MasterAuthJSON)
			}
			if clusterData.AutoscalingJSON != nil {
				create.SetAutoscalingJSON(clusterData.AutoscalingJSON)
			}
			if clusterData.VerticalPodAutoscalingJSON != nil {
				create.SetVerticalPodAutoscalingJSON(clusterData.VerticalPodAutoscalingJSON)
			}
			if clusterData.MonitoringConfigJSON != nil {
				create.SetMonitoringConfigJSON(clusterData.MonitoringConfigJSON)
			}
			if clusterData.LoggingConfigJSON != nil {
				create.SetLoggingConfigJSON(clusterData.LoggingConfigJSON)
			}
			if clusterData.MaintenancePolicyJSON != nil {
				create.SetMaintenancePolicyJSON(clusterData.MaintenancePolicyJSON)
			}
			if clusterData.DatabaseEncryptionJSON != nil {
				create.SetDatabaseEncryptionJSON(clusterData.DatabaseEncryptionJSON)
			}
			if clusterData.WorkloadIdentityConfigJSON != nil {
				create.SetWorkloadIdentityConfigJSON(clusterData.WorkloadIdentityConfigJSON)
			}
			if clusterData.AutopilotJSON != nil {
				create.SetAutopilotJSON(clusterData.AutopilotJSON)
			}
			if clusterData.ReleaseChannelJSON != nil {
				create.SetReleaseChannelJSON(clusterData.ReleaseChannelJSON)
			}
			if clusterData.BinaryAuthorizationJSON != nil {
				create.SetBinaryAuthorizationJSON(clusterData.BinaryAuthorizationJSON)
			}
			if clusterData.SecurityPostureConfigJSON != nil {
				create.SetSecurityPostureConfigJSON(clusterData.SecurityPostureConfigJSON)
			}
			if clusterData.NodePoolDefaultsJSON != nil {
				create.SetNodePoolDefaultsJSON(clusterData.NodePoolDefaultsJSON)
			}
			if clusterData.FleetJSON != nil {
				create.SetFleetJSON(clusterData.FleetJSON)
			}

			savedCluster, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create cluster %s: %w", clusterData.Name, err)
			}
		} else {
			// Update existing cluster
			update := tx.BronzeGCPContainerCluster.UpdateOneID(clusterData.ResourceID).
				SetName(clusterData.Name).
				SetLocation(clusterData.Location).
				SetZone(clusterData.Zone).
				SetDescription(clusterData.Description).
				SetInitialClusterVersion(clusterData.InitialClusterVersion).
				SetCurrentMasterVersion(clusterData.CurrentMasterVersion).
				SetCurrentNodeVersion(clusterData.CurrentNodeVersion).
				SetStatus(clusterData.Status).
				SetStatusMessage(clusterData.StatusMessage).
				SetCurrentNodeCount(clusterData.CurrentNodeCount).
				SetNetwork(clusterData.Network).
				SetSubnetwork(clusterData.Subnetwork).
				SetClusterIpv4Cidr(clusterData.ClusterIpv4Cidr).
				SetServicesIpv4Cidr(clusterData.ServicesIpv4Cidr).
				SetNodeIpv4CidrSize(clusterData.NodeIpv4CidrSize).
				SetEndpoint(clusterData.Endpoint).
				SetSelfLink(clusterData.SelfLink).
				SetCreateTime(clusterData.CreateTime).
				SetExpireTime(clusterData.ExpireTime).
				SetEtag(clusterData.Etag).
				SetLabelFingerprint(clusterData.LabelFingerprint).
				SetLoggingService(clusterData.LoggingService).
				SetMonitoringService(clusterData.MonitoringService).
				SetEnableKubernetesAlpha(clusterData.EnableKubernetesAlpha).
				SetEnableTpu(clusterData.EnableTpu).
				SetTpuIpv4CidrBlock(clusterData.TpuIpv4CidrBlock).
				SetProjectID(clusterData.ProjectID).
				SetCollectedAt(clusterData.CollectedAt)

			// Set optional JSON fields
			if clusterData.AddonsConfigJSON != nil {
				update.SetAddonsConfigJSON(clusterData.AddonsConfigJSON)
			}
			if clusterData.PrivateClusterConfigJSON != nil {
				update.SetPrivateClusterConfigJSON(clusterData.PrivateClusterConfigJSON)
			}
			if clusterData.IPAllocationPolicyJSON != nil {
				update.SetIPAllocationPolicyJSON(clusterData.IPAllocationPolicyJSON)
			}
			if clusterData.NetworkConfigJSON != nil {
				update.SetNetworkConfigJSON(clusterData.NetworkConfigJSON)
			}
			if clusterData.MasterAuthJSON != nil {
				update.SetMasterAuthJSON(clusterData.MasterAuthJSON)
			}
			if clusterData.AutoscalingJSON != nil {
				update.SetAutoscalingJSON(clusterData.AutoscalingJSON)
			}
			if clusterData.VerticalPodAutoscalingJSON != nil {
				update.SetVerticalPodAutoscalingJSON(clusterData.VerticalPodAutoscalingJSON)
			}
			if clusterData.MonitoringConfigJSON != nil {
				update.SetMonitoringConfigJSON(clusterData.MonitoringConfigJSON)
			}
			if clusterData.LoggingConfigJSON != nil {
				update.SetLoggingConfigJSON(clusterData.LoggingConfigJSON)
			}
			if clusterData.MaintenancePolicyJSON != nil {
				update.SetMaintenancePolicyJSON(clusterData.MaintenancePolicyJSON)
			}
			if clusterData.DatabaseEncryptionJSON != nil {
				update.SetDatabaseEncryptionJSON(clusterData.DatabaseEncryptionJSON)
			}
			if clusterData.WorkloadIdentityConfigJSON != nil {
				update.SetWorkloadIdentityConfigJSON(clusterData.WorkloadIdentityConfigJSON)
			}
			if clusterData.AutopilotJSON != nil {
				update.SetAutopilotJSON(clusterData.AutopilotJSON)
			}
			if clusterData.ReleaseChannelJSON != nil {
				update.SetReleaseChannelJSON(clusterData.ReleaseChannelJSON)
			}
			if clusterData.BinaryAuthorizationJSON != nil {
				update.SetBinaryAuthorizationJSON(clusterData.BinaryAuthorizationJSON)
			}
			if clusterData.SecurityPostureConfigJSON != nil {
				update.SetSecurityPostureConfigJSON(clusterData.SecurityPostureConfigJSON)
			}
			if clusterData.NodePoolDefaultsJSON != nil {
				update.SetNodePoolDefaultsJSON(clusterData.NodePoolDefaultsJSON)
			}
			if clusterData.FleetJSON != nil {
				update.SetFleetJSON(clusterData.FleetJSON)
			}

			savedCluster, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update cluster %s: %w", clusterData.Name, err)
			}
		}

		// Create child entities
		if err := s.createClusterChildren(ctx, tx, savedCluster, clusterData); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create children for cluster %s: %w", clusterData.Name, err)
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, clusterData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for cluster %s: %w", clusterData.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, clusterData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for cluster %s: %w", clusterData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// deleteClusterChildren deletes all child entities for a cluster.
func (s *Service) deleteClusterChildren(ctx context.Context, tx *ent.Tx, clusterID string) error {
	// Labels
	_, err := tx.BronzeGCPContainerClusterLabel.Delete().
		Where(bronzegcpcontainerclusterlabel.HasClusterWith(bronzegcpcontainercluster.ID(clusterID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete labels: %w", err)
	}

	// Addons
	_, err = tx.BronzeGCPContainerClusterAddon.Delete().
		Where(bronzegcpcontainerclusteraddon.HasClusterWith(bronzegcpcontainercluster.ID(clusterID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete addons: %w", err)
	}

	// Conditions
	_, err = tx.BronzeGCPContainerClusterCondition.Delete().
		Where(bronzegcpcontainerclustercondition.HasClusterWith(bronzegcpcontainercluster.ID(clusterID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete conditions: %w", err)
	}

	// Node pools
	_, err = tx.BronzeGCPContainerClusterNodePool.Delete().
		Where(bronzegcpcontainerclusternodepool.HasClusterWith(bronzegcpcontainercluster.ID(clusterID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete node pools: %w", err)
	}

	return nil
}

// createClusterChildren creates all child entities for a cluster.
func (s *Service) createClusterChildren(ctx context.Context, tx *ent.Tx, cluster *ent.BronzeGCPContainerCluster, data *ClusterData) error {
	// Create labels
	for _, labelData := range data.Labels {
		_, err := tx.BronzeGCPContainerClusterLabel.Create().
			SetKey(labelData.Key).
			SetValue(labelData.Value).
			SetCluster(cluster).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label: %w", err)
		}
	}

	// Create addons
	for _, addonData := range data.Addons {
		create := tx.BronzeGCPContainerClusterAddon.Create().
			SetAddonName(addonData.AddonName).
			SetEnabled(addonData.Enabled).
			SetCluster(cluster)

		if addonData.ConfigJSON != nil {
			create.SetConfigJSON(addonData.ConfigJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create addon: %w", err)
		}
	}

	// Create conditions
	for _, condData := range data.Conditions {
		_, err := tx.BronzeGCPContainerClusterCondition.Create().
			SetCode(condData.Code).
			SetMessage(condData.Message).
			SetCanonicalCode(condData.CanonicalCode).
			SetCluster(cluster).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create condition: %w", err)
		}
	}

	// Create node pools
	for _, npData := range data.NodePools {
		create := tx.BronzeGCPContainerClusterNodePool.Create().
			SetName(npData.Name).
			SetVersion(npData.Version).
			SetStatus(npData.Status).
			SetStatusMessage(npData.StatusMessage).
			SetInitialNodeCount(npData.InitialNodeCount).
			SetSelfLink(npData.SelfLink).
			SetPodIpv4CidrSize(npData.PodIpv4CidrSize).
			SetEtag(npData.Etag).
			SetCluster(cluster)

		if npData.LocationsJSON != nil {
			create.SetLocationsJSON(npData.LocationsJSON)
		}
		if npData.ConfigJSON != nil {
			create.SetConfigJSON(npData.ConfigJSON)
		}
		if npData.AutoscalingJSON != nil {
			create.SetAutoscalingJSON(npData.AutoscalingJSON)
		}
		if npData.ManagementJSON != nil {
			create.SetManagementJSON(npData.ManagementJSON)
		}
		if npData.UpgradeSettingsJSON != nil {
			create.SetUpgradeSettingsJSON(npData.UpgradeSettingsJSON)
		}
		if npData.NetworkConfigJSON != nil {
			create.SetNetworkConfigJSON(npData.NetworkConfigJSON)
		}
		if npData.PlacementPolicyJSON != nil {
			create.SetPlacementPolicyJSON(npData.PlacementPolicyJSON)
		}
		if npData.MaxPodsConstraintJSON != nil {
			create.SetMaxPodsConstraintJSON(npData.MaxPodsConstraintJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create node pool: %w", err)
		}
	}

	return nil
}

// DeleteStaleClusters removes clusters that were not collected in the latest run.
// Also closes history records for deleted clusters.
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

	// Find stale clusters
	staleClusters, err := tx.BronzeGCPContainerCluster.Query().
		Where(
			bronzegcpcontainercluster.ProjectID(projectID),
			bronzegcpcontainercluster.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to find stale clusters: %w", err)
	}

	// Close history and delete each stale cluster
	for _, cluster := range staleClusters {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, cluster.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for cluster %s: %w", cluster.ID, err)
		}

		// Delete children (CASCADE DELETE will handle this, but we do it explicitly)
		if err := s.deleteClusterChildren(ctx, tx, cluster.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete children for cluster %s: %w", cluster.ID, err)
		}

		// Delete cluster
		if err := tx.BronzeGCPContainerCluster.DeleteOne(cluster).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete cluster %s: %w", cluster.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
