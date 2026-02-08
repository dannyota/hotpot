package cluster

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/container/apiv1/containerpb"
)

// ClusterData holds converted cluster data ready for Ent insertion.
type ClusterData struct {
	ResourceID            string
	Name                  string
	Location              string
	Zone                  string
	Description           string
	InitialClusterVersion string
	CurrentMasterVersion  string
	CurrentNodeVersion    string
	Status                string
	StatusMessage         string
	CurrentNodeCount      int32
	Network               string
	Subnetwork            string
	ClusterIpv4Cidr       string
	ServicesIpv4Cidr      string
	NodeIpv4CidrSize      int32
	Endpoint              string
	SelfLink              string
	CreateTime            string
	ExpireTime            string
	Etag                  string
	LabelFingerprint      string
	LoggingService        string
	MonitoringService     string
	EnableKubernetesAlpha bool
	EnableTpu             bool
	TpuIpv4CidrBlock      string
	ProjectID             string
	CollectedAt           time.Time

	// Nested JSON fields
	AddonsConfigJSON           json.RawMessage
	PrivateClusterConfigJSON   json.RawMessage
	IPAllocationPolicyJSON     json.RawMessage
	NetworkConfigJSON          json.RawMessage
	MasterAuthJSON             json.RawMessage
	AutoscalingJSON            json.RawMessage
	VerticalPodAutoscalingJSON json.RawMessage
	MonitoringConfigJSON       json.RawMessage
	LoggingConfigJSON          json.RawMessage
	MaintenancePolicyJSON      json.RawMessage
	DatabaseEncryptionJSON     json.RawMessage
	WorkloadIdentityConfigJSON json.RawMessage
	AutopilotJSON              json.RawMessage
	ReleaseChannelJSON         json.RawMessage
	BinaryAuthorizationJSON    json.RawMessage
	SecurityPostureConfigJSON  json.RawMessage
	NodePoolDefaultsJSON       json.RawMessage
	FleetJSON                  json.RawMessage

	// Child data
	Labels     []LabelData
	Addons     []AddonData
	Conditions []ConditionData
	NodePools  []NodePoolData
}

type LabelData struct {
	Key   string
	Value string
}

type AddonData struct {
	AddonName  string
	Enabled    bool
	ConfigJSON json.RawMessage
}

type ConditionData struct {
	Code          string
	Message       string
	CanonicalCode string
}

type NodePoolData struct {
	Name                  string
	Version               string
	Status                string
	StatusMessage         string
	InitialNodeCount      int32
	SelfLink              string
	PodIpv4CidrSize       int32
	Etag                  string
	LocationsJSON         json.RawMessage
	ConfigJSON            json.RawMessage
	AutoscalingJSON       json.RawMessage
	ManagementJSON        json.RawMessage
	UpgradeSettingsJSON   json.RawMessage
	NetworkConfigJSON     json.RawMessage
	PlacementPolicyJSON   json.RawMessage
	MaxPodsConstraintJSON json.RawMessage
}

// ConvertCluster converts a GCP API Cluster to ClusterData.
// Preserves raw API data with minimal transformation.
func ConvertCluster(cluster *containerpb.Cluster, projectID string, collectedAt time.Time) (*ClusterData, error) {
	c := &ClusterData{
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

	// Convert nested objects to JSONB (nil → SQL NULL, data → JSON bytes)
	var err error
	if cluster.AddonsConfig != nil {
		c.AddonsConfigJSON, err = json.Marshal(cluster.AddonsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for cluster %s: %w", cluster.GetName(), err)
		}
	}
	if cluster.PrivateClusterConfig != nil {
		c.PrivateClusterConfigJSON, err = json.Marshal(cluster.PrivateClusterConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for cluster %s: %w", cluster.GetName(), err)
		}
	}
	if cluster.IpAllocationPolicy != nil {
		c.IPAllocationPolicyJSON, err = json.Marshal(cluster.IpAllocationPolicy)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for cluster %s: %w", cluster.GetName(), err)
		}
	}
	if cluster.NetworkConfig != nil {
		c.NetworkConfigJSON, err = json.Marshal(cluster.NetworkConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for cluster %s: %w", cluster.GetName(), err)
		}
	}
	if cluster.MasterAuth != nil {
		c.MasterAuthJSON, err = json.Marshal(cluster.MasterAuth)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for cluster %s: %w", cluster.GetName(), err)
		}
	}
	if cluster.Autoscaling != nil {
		c.AutoscalingJSON, err = json.Marshal(cluster.Autoscaling)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for cluster %s: %w", cluster.GetName(), err)
		}
	}
	if cluster.VerticalPodAutoscaling != nil {
		c.VerticalPodAutoscalingJSON, err = json.Marshal(cluster.VerticalPodAutoscaling)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for cluster %s: %w", cluster.GetName(), err)
		}
	}
	if cluster.MonitoringConfig != nil {
		c.MonitoringConfigJSON, err = json.Marshal(cluster.MonitoringConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for cluster %s: %w", cluster.GetName(), err)
		}
	}
	if cluster.LoggingConfig != nil {
		c.LoggingConfigJSON, err = json.Marshal(cluster.LoggingConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for cluster %s: %w", cluster.GetName(), err)
		}
	}
	if cluster.MaintenancePolicy != nil {
		c.MaintenancePolicyJSON, err = json.Marshal(cluster.MaintenancePolicy)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for cluster %s: %w", cluster.GetName(), err)
		}
	}
	if cluster.DatabaseEncryption != nil {
		c.DatabaseEncryptionJSON, err = json.Marshal(cluster.DatabaseEncryption)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for cluster %s: %w", cluster.GetName(), err)
		}
	}
	if cluster.WorkloadIdentityConfig != nil {
		c.WorkloadIdentityConfigJSON, err = json.Marshal(cluster.WorkloadIdentityConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for cluster %s: %w", cluster.GetName(), err)
		}
	}
	if cluster.Autopilot != nil {
		c.AutopilotJSON, err = json.Marshal(cluster.Autopilot)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for cluster %s: %w", cluster.GetName(), err)
		}
	}
	if cluster.ReleaseChannel != nil {
		c.ReleaseChannelJSON, err = json.Marshal(cluster.ReleaseChannel)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for cluster %s: %w", cluster.GetName(), err)
		}
	}
	if cluster.BinaryAuthorization != nil {
		c.BinaryAuthorizationJSON, err = json.Marshal(cluster.BinaryAuthorization)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for cluster %s: %w", cluster.GetName(), err)
		}
	}
	if cluster.SecurityPostureConfig != nil {
		c.SecurityPostureConfigJSON, err = json.Marshal(cluster.SecurityPostureConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for cluster %s: %w", cluster.GetName(), err)
		}
	}
	if cluster.NodePoolDefaults != nil {
		c.NodePoolDefaultsJSON, err = json.Marshal(cluster.NodePoolDefaults)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for cluster %s: %w", cluster.GetName(), err)
		}
	}
	if cluster.Fleet != nil {
		c.FleetJSON, err = json.Marshal(cluster.Fleet)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal JSON for cluster %s: %w", cluster.GetName(), err)
		}
	}

	// Convert related entities to separate tables
	c.Labels = ConvertLabels(cluster.GetResourceLabels())
	c.Addons, err = ConvertAddons(cluster.GetAddonsConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to convert addons for cluster %s: %w", cluster.GetName(), err)
	}
	c.Conditions = ConvertConditions(cluster.GetConditions())
	c.NodePools, err = ConvertNodePools(cluster.GetNodePools())
	if err != nil {
		return nil, fmt.Errorf("failed to convert node pools for cluster %s: %w", cluster.GetName(), err)
	}

	return c, nil
}

// ConvertLabels converts cluster resource labels to data structs.
func ConvertLabels(labels map[string]string) []LabelData {
	if len(labels) == 0 {
		return nil
	}

	result := make([]LabelData, 0, len(labels))
	for key, value := range labels {
		result = append(result, LabelData{
			Key:   key,
			Value: value,
		})
	}

	return result
}

// marshalAddonConfig marshals a config and returns the addon entry.
// cfg is already known to be non-nil when called.
func marshalAddonConfig(name string, enabled bool, cfg any) (AddonData, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return AddonData{}, fmt.Errorf("failed to marshal %s config: %w", name, err)
	}
	return AddonData{
		AddonName:  name,
		Enabled:    enabled,
		ConfigJSON: data,
	}, nil
}

// ConvertAddons converts addons config to data structs (one row per addon).
func ConvertAddons(addonsConfig *containerpb.AddonsConfig) ([]AddonData, error) {
	if addonsConfig == nil {
		return nil, nil
	}

	var result []AddonData
	add := func(name string, enabled bool, cfg any) error {
		addon, err := marshalAddonConfig(name, enabled, cfg)
		if err != nil {
			return err
		}
		result = append(result, addon)
		return nil
	}

	if cfg := addonsConfig.GetHttpLoadBalancing(); cfg != nil {
		if err := add("http_load_balancing", !cfg.GetDisabled(), cfg); err != nil {
			return nil, err
		}
	}
	if cfg := addonsConfig.GetHorizontalPodAutoscaling(); cfg != nil {
		if err := add("horizontal_pod_autoscaling", !cfg.GetDisabled(), cfg); err != nil {
			return nil, err
		}
	}
	if cfg := addonsConfig.GetKubernetesDashboard(); cfg != nil {
		if err := add("kubernetes_dashboard", !cfg.GetDisabled(), cfg); err != nil {
			return nil, err
		}
	}
	if cfg := addonsConfig.GetNetworkPolicyConfig(); cfg != nil {
		if err := add("network_policy", !cfg.GetDisabled(), cfg); err != nil {
			return nil, err
		}
	}
	if cfg := addonsConfig.GetCloudRunConfig(); cfg != nil {
		if err := add("cloud_run", !cfg.GetDisabled(), cfg); err != nil {
			return nil, err
		}
	}
	if cfg := addonsConfig.GetDnsCacheConfig(); cfg != nil {
		if err := add("dns_cache", cfg.GetEnabled(), cfg); err != nil {
			return nil, err
		}
	}
	if cfg := addonsConfig.GetConfigConnectorConfig(); cfg != nil {
		if err := add("config_connector", cfg.GetEnabled(), cfg); err != nil {
			return nil, err
		}
	}
	if cfg := addonsConfig.GetGcePersistentDiskCsiDriverConfig(); cfg != nil {
		if err := add("gce_persistent_disk_csi_driver", cfg.GetEnabled(), cfg); err != nil {
			return nil, err
		}
	}
	if cfg := addonsConfig.GetGcpFilestoreCsiDriverConfig(); cfg != nil {
		if err := add("gcp_filestore_csi_driver", cfg.GetEnabled(), cfg); err != nil {
			return nil, err
		}
	}
	if cfg := addonsConfig.GetGcsFuseCsiDriverConfig(); cfg != nil {
		if err := add("gcs_fuse_csi_driver", cfg.GetEnabled(), cfg); err != nil {
			return nil, err
		}
	}
	if cfg := addonsConfig.GetGkeBackupAgentConfig(); cfg != nil {
		if err := add("gke_backup_agent", cfg.GetEnabled(), cfg); err != nil {
			return nil, err
		}
	}
	if cfg := addonsConfig.GetRayOperatorConfig(); cfg != nil {
		if err := add("ray_operator", cfg.GetEnabled(), cfg); err != nil {
			return nil, err
		}
	}
	if cfg := addonsConfig.GetStatefulHaConfig(); cfg != nil {
		if err := add("stateful_ha", cfg.GetEnabled(), cfg); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// ConvertConditions converts cluster conditions to data structs.
func ConvertConditions(conditions []*containerpb.StatusCondition) []ConditionData {
	if len(conditions) == 0 {
		return nil
	}

	result := make([]ConditionData, 0, len(conditions))
	for _, cond := range conditions {
		result = append(result, ConditionData{
			Code:          cond.GetCode().String(),
			Message:       cond.GetMessage(),
			CanonicalCode: cond.GetCanonicalCode().String(),
		})
	}

	return result
}

// ConvertNodePools converts node pools to data structs.
func ConvertNodePools(nodePools []*containerpb.NodePool) ([]NodePoolData, error) {
	if len(nodePools) == 0 {
		return nil, nil
	}

	result := make([]NodePoolData, 0, len(nodePools))
	for _, np := range nodePools {
		pool := NodePoolData{
			Name:             np.GetName(),
			Version:          np.GetVersion(),
			Status:           np.GetStatus().String(),
			StatusMessage:    np.GetStatusMessage(),
			InitialNodeCount: np.GetInitialNodeCount(),
			SelfLink:         np.GetSelfLink(),
			PodIpv4CidrSize:  np.GetPodIpv4CidrSize(),
			Etag:             np.GetEtag(),
		}

		// Convert nested objects to JSONB (nil → SQL NULL, data → JSON bytes)
		var err error
		if np.Locations != nil {
			pool.LocationsJSON, err = json.Marshal(np.Locations)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal node pool %s JSON: %w", np.GetName(), err)
			}
		}
		if np.Config != nil {
			pool.ConfigJSON, err = json.Marshal(np.Config)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal node pool %s JSON: %w", np.GetName(), err)
			}
		}
		if np.Autoscaling != nil {
			pool.AutoscalingJSON, err = json.Marshal(np.Autoscaling)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal node pool %s JSON: %w", np.GetName(), err)
			}
		}
		if np.Management != nil {
			pool.ManagementJSON, err = json.Marshal(np.Management)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal node pool %s JSON: %w", np.GetName(), err)
			}
		}
		if np.UpgradeSettings != nil {
			pool.UpgradeSettingsJSON, err = json.Marshal(np.UpgradeSettings)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal node pool %s JSON: %w", np.GetName(), err)
			}
		}
		if np.NetworkConfig != nil {
			pool.NetworkConfigJSON, err = json.Marshal(np.NetworkConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal node pool %s JSON: %w", np.GetName(), err)
			}
		}
		if np.PlacementPolicy != nil {
			pool.PlacementPolicyJSON, err = json.Marshal(np.PlacementPolicy)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal node pool %s JSON: %w", np.GetName(), err)
			}
		}
		if np.MaxPodsConstraint != nil {
			pool.MaxPodsConstraintJSON, err = json.Marshal(np.MaxPodsConstraint)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal node pool %s JSON: %w", np.GetName(), err)
			}
		}

		result = append(result, pool)
	}

	return result, nil
}
