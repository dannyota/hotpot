package cluster

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/container/apiv1/containerpb"

	"hotpot/pkg/base/models/bronze"
)

// ConvertCluster converts a GCP API Cluster to a Bronze model.
// Preserves raw API data with minimal transformation.
func ConvertCluster(cluster *containerpb.Cluster, projectID string, collectedAt time.Time) bronze.GCPContainerCluster {
	c := bronze.GCPContainerCluster{
		ResourceID:            cluster.GetId(),
		Name:                  cluster.GetName(),
		Location:              cluster.GetLocation(),
		Zone:                  cluster.GetZone(),
		Description:           cluster.GetDescription(),
		InitialClusterVersion: cluster.GetInitialClusterVersion(),
		CurrentMasterVersion:  cluster.GetCurrentMasterVersion(),
		CurrentNodeVersion:    cluster.GetCurrentNodeVersion(),
		Status:                cluster.GetStatus().String(),
		StatusMessage:         cluster.GetStatusMessage(),
		CurrentNodeCount:      cluster.GetCurrentNodeCount(),
		Network:               cluster.GetNetwork(),
		Subnetwork:            cluster.GetSubnetwork(),
		ClusterIpv4Cidr:       cluster.GetClusterIpv4Cidr(),
		ServicesIpv4Cidr:      cluster.GetServicesIpv4Cidr(),
		NodeIpv4CidrSize:      cluster.GetNodeIpv4CidrSize(),
		Endpoint:              cluster.GetEndpoint(),
		SelfLink:              cluster.GetSelfLink(),
		CreateTime:            cluster.GetCreateTime(),
		ExpireTime:            cluster.GetExpireTime(),
		Etag:                  cluster.GetEtag(),
		LabelFingerprint:      cluster.GetLabelFingerprint(),
		LoggingService:        cluster.GetLoggingService(),
		MonitoringService:     cluster.GetMonitoringService(),
		EnableKubernetesAlpha: cluster.GetEnableKubernetesAlpha(),
		EnableTpu:             cluster.GetEnableTpu(),
		TpuIpv4CidrBlock:      cluster.GetTpuIpv4CidrBlock(),
		ProjectID:             projectID,
		CollectedAt:           collectedAt,
	}

	// Convert nested objects to JSONB
	c.AddonsConfigJSON = toJSON(cluster.GetAddonsConfig())
	c.PrivateClusterConfigJSON = toJSON(cluster.GetPrivateClusterConfig())
	c.IpAllocationPolicyJSON = toJSON(cluster.GetIpAllocationPolicy())
	c.NetworkConfigJSON = toJSON(cluster.GetNetworkConfig())
	c.MasterAuthJSON = toJSON(cluster.GetMasterAuth())
	c.AutoscalingJSON = toJSON(cluster.GetAutoscaling())
	c.VerticalPodAutoscalingJSON = toJSON(cluster.GetVerticalPodAutoscaling())
	c.MonitoringConfigJSON = toJSON(cluster.GetMonitoringConfig())
	c.LoggingConfigJSON = toJSON(cluster.GetLoggingConfig())
	c.MaintenancePolicyJSON = toJSON(cluster.GetMaintenancePolicy())
	c.DatabaseEncryptionJSON = toJSON(cluster.GetDatabaseEncryption())
	c.WorkloadIdentityConfigJSON = toJSON(cluster.GetWorkloadIdentityConfig())
	c.AutopilotJSON = toJSON(cluster.GetAutopilot())
	c.ReleaseChannelJSON = toJSON(cluster.GetReleaseChannel())
	c.BinaryAuthorizationJSON = toJSON(cluster.GetBinaryAuthorization())
	c.SecurityPostureConfigJSON = toJSON(cluster.GetSecurityPostureConfig())
	c.NodePoolDefaultsJSON = toJSON(cluster.GetNodePoolDefaults())
	c.FleetJSON = toJSON(cluster.GetFleet())

	// Convert related entities to separate tables
	c.Labels = ConvertLabels(cluster.GetResourceLabels())
	c.Addons = ConvertAddons(cluster.GetAddonsConfig())
	c.Conditions = ConvertConditions(cluster.GetConditions())
	c.NodePools = ConvertNodePools(cluster.GetNodePools())

	return c
}

// ConvertLabels converts cluster resource labels to Bronze models.
func ConvertLabels(labels map[string]string) []bronze.GCPContainerClusterLabel {
	if len(labels) == 0 {
		return nil
	}

	result := make([]bronze.GCPContainerClusterLabel, 0, len(labels))
	for key, value := range labels {
		result = append(result, bronze.GCPContainerClusterLabel{
			Key:   key,
			Value: value,
		})
	}

	return result
}

// ConvertAddons converts addons config to Bronze models (one row per addon).
func ConvertAddons(addonsConfig *containerpb.AddonsConfig) []bronze.GCPContainerClusterAddon {
	if addonsConfig == nil {
		return nil
	}

	var result []bronze.GCPContainerClusterAddon

	// HTTP Load Balancing
	if cfg := addonsConfig.GetHttpLoadBalancing(); cfg != nil {
		result = append(result, bronze.GCPContainerClusterAddon{
			AddonName:  "http_load_balancing",
			Enabled:    !cfg.GetDisabled(),
			ConfigJSON: toJSON(cfg),
		})
	}

	// Horizontal Pod Autoscaling
	if cfg := addonsConfig.GetHorizontalPodAutoscaling(); cfg != nil {
		result = append(result, bronze.GCPContainerClusterAddon{
			AddonName:  "horizontal_pod_autoscaling",
			Enabled:    !cfg.GetDisabled(),
			ConfigJSON: toJSON(cfg),
		})
	}

	// Kubernetes Dashboard (deprecated)
	if cfg := addonsConfig.GetKubernetesDashboard(); cfg != nil {
		result = append(result, bronze.GCPContainerClusterAddon{
			AddonName:  "kubernetes_dashboard",
			Enabled:    !cfg.GetDisabled(),
			ConfigJSON: toJSON(cfg),
		})
	}

	// Network Policy Config
	if cfg := addonsConfig.GetNetworkPolicyConfig(); cfg != nil {
		result = append(result, bronze.GCPContainerClusterAddon{
			AddonName:  "network_policy",
			Enabled:    !cfg.GetDisabled(),
			ConfigJSON: toJSON(cfg),
		})
	}

	// Cloud Run Config
	if cfg := addonsConfig.GetCloudRunConfig(); cfg != nil {
		result = append(result, bronze.GCPContainerClusterAddon{
			AddonName:  "cloud_run",
			Enabled:    !cfg.GetDisabled(),
			ConfigJSON: toJSON(cfg),
		})
	}

	// DNS Cache Config
	if cfg := addonsConfig.GetDnsCacheConfig(); cfg != nil {
		result = append(result, bronze.GCPContainerClusterAddon{
			AddonName:  "dns_cache",
			Enabled:    cfg.GetEnabled(),
			ConfigJSON: toJSON(cfg),
		})
	}

	// Config Connector Config
	if cfg := addonsConfig.GetConfigConnectorConfig(); cfg != nil {
		result = append(result, bronze.GCPContainerClusterAddon{
			AddonName:  "config_connector",
			Enabled:    cfg.GetEnabled(),
			ConfigJSON: toJSON(cfg),
		})
	}

	// GCE Persistent Disk CSI Driver Config
	if cfg := addonsConfig.GetGcePersistentDiskCsiDriverConfig(); cfg != nil {
		result = append(result, bronze.GCPContainerClusterAddon{
			AddonName:  "gce_persistent_disk_csi_driver",
			Enabled:    cfg.GetEnabled(),
			ConfigJSON: toJSON(cfg),
		})
	}

	// GCP Filestore CSI Driver Config
	if cfg := addonsConfig.GetGcpFilestoreCsiDriverConfig(); cfg != nil {
		result = append(result, bronze.GCPContainerClusterAddon{
			AddonName:  "gcp_filestore_csi_driver",
			Enabled:    cfg.GetEnabled(),
			ConfigJSON: toJSON(cfg),
		})
	}

	// GCS Fuse CSI Driver Config
	if cfg := addonsConfig.GetGcsFuseCsiDriverConfig(); cfg != nil {
		result = append(result, bronze.GCPContainerClusterAddon{
			AddonName:  "gcs_fuse_csi_driver",
			Enabled:    cfg.GetEnabled(),
			ConfigJSON: toJSON(cfg),
		})
	}

	// GKE Backup Agent Config
	if cfg := addonsConfig.GetGkeBackupAgentConfig(); cfg != nil {
		result = append(result, bronze.GCPContainerClusterAddon{
			AddonName:  "gke_backup_agent",
			Enabled:    cfg.GetEnabled(),
			ConfigJSON: toJSON(cfg),
		})
	}

	// Ray Operator Config
	if cfg := addonsConfig.GetRayOperatorConfig(); cfg != nil {
		result = append(result, bronze.GCPContainerClusterAddon{
			AddonName:  "ray_operator",
			Enabled:    cfg.GetEnabled(),
			ConfigJSON: toJSON(cfg),
		})
	}

	// Stateful HA Config
	if cfg := addonsConfig.GetStatefulHaConfig(); cfg != nil {
		result = append(result, bronze.GCPContainerClusterAddon{
			AddonName:  "stateful_ha",
			Enabled:    cfg.GetEnabled(),
			ConfigJSON: toJSON(cfg),
		})
	}

	return result
}

// ConvertConditions converts cluster conditions to Bronze models.
func ConvertConditions(conditions []*containerpb.StatusCondition) []bronze.GCPContainerClusterCondition {
	if len(conditions) == 0 {
		return nil
	}

	result := make([]bronze.GCPContainerClusterCondition, 0, len(conditions))
	for _, cond := range conditions {
		result = append(result, bronze.GCPContainerClusterCondition{
			Code:          cond.GetCode().String(),
			Message:       cond.GetMessage(),
			CanonicalCode: cond.GetCanonicalCode().String(),
		})
	}

	return result
}

// ConvertNodePools converts node pools to Bronze models.
func ConvertNodePools(nodePools []*containerpb.NodePool) []bronze.GCPContainerClusterNodePool {
	if len(nodePools) == 0 {
		return nil
	}

	result := make([]bronze.GCPContainerClusterNodePool, 0, len(nodePools))
	for _, np := range nodePools {
		pool := bronze.GCPContainerClusterNodePool{
			Name:             np.GetName(),
			Version:          np.GetVersion(),
			Status:           np.GetStatus().String(),
			StatusMessage:    np.GetStatusMessage(),
			InitialNodeCount: np.GetInitialNodeCount(),
			SelfLink:         np.GetSelfLink(),
			PodIpv4CidrSize:  np.GetPodIpv4CidrSize(),
			Etag:             np.GetEtag(),
		}

		// Convert nested objects to JSONB
		pool.LocationsJSON = toJSON(np.GetLocations())
		pool.ConfigJSON = toJSON(np.GetConfig())
		pool.AutoscalingJSON = toJSON(np.GetAutoscaling())
		pool.ManagementJSON = toJSON(np.GetManagement())
		pool.UpgradeSettingsJSON = toJSON(np.GetUpgradeSettings())
		pool.NetworkConfigJSON = toJSON(np.GetNetworkConfig())
		pool.PlacementPolicyJSON = toJSON(np.GetPlacementPolicy())
		pool.MaxPodsConstraintJSON = toJSON(np.GetMaxPodsConstraint())

		result = append(result, pool)
	}

	return result
}

// toJSON converts any value to JSON string, returns empty string on error.
func toJSON(v any) string {
	if v == nil {
		return ""
	}
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(data)
}
