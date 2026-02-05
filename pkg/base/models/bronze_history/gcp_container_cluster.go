package bronze_history

import (
	"time"
)

// GCPContainerCluster stores historical snapshots of GKE clusters.
// Uses ResourceID for lookup (has API ID), with valid_from/valid_to for time range.
type GCPContainerCluster struct {
	HistoryID uint `gorm:"primaryKey"`

	// Link by API ID
	ResourceID string `gorm:"column:resource_id;type:varchar(255);not null;index" json:"id"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"` // null = current

	// All cluster fields (same as bronze.GCPContainerCluster)
	Name                       string    `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Location                   string    `gorm:"column:location;type:varchar(100)" json:"location"`
	Zone                       string    `gorm:"column:zone;type:varchar(100)" json:"zone"`
	Description                string    `gorm:"column:description;type:text" json:"description"`
	InitialClusterVersion      string    `gorm:"column:initial_cluster_version;type:varchar(50)" json:"initialClusterVersion"`
	CurrentMasterVersion       string    `gorm:"column:current_master_version;type:varchar(50)" json:"currentMasterVersion"`
	CurrentNodeVersion         string    `gorm:"column:current_node_version;type:varchar(50)" json:"currentNodeVersion"`
	Status                     string    `gorm:"column:status;type:varchar(50)" json:"status"`
	StatusMessage              string    `gorm:"column:status_message;type:text" json:"statusMessage"`
	CurrentNodeCount           int32     `gorm:"column:current_node_count" json:"currentNodeCount"`
	Network                    string    `gorm:"column:network;type:varchar(255)" json:"network"`
	Subnetwork                 string    `gorm:"column:subnetwork;type:varchar(255)" json:"subnetwork"`
	ClusterIpv4Cidr            string    `gorm:"column:cluster_ipv4_cidr;type:varchar(50)" json:"clusterIpv4Cidr"`
	ServicesIpv4Cidr           string    `gorm:"column:services_ipv4_cidr;type:varchar(50)" json:"servicesIpv4Cidr"`
	NodeIpv4CidrSize           int32     `gorm:"column:node_ipv4_cidr_size" json:"nodeIpv4CidrSize"`
	Endpoint                   string    `gorm:"column:endpoint;type:varchar(255)" json:"endpoint"`
	SelfLink                   string    `gorm:"column:self_link;type:text" json:"selfLink"`
	CreateTime                 string    `gorm:"column:create_time;type:varchar(50)" json:"createTime"`
	ExpireTime                 string    `gorm:"column:expire_time;type:varchar(50)" json:"expireTime"`
	Etag                       string    `gorm:"column:etag;type:varchar(255)" json:"etag"`
	LabelFingerprint           string    `gorm:"column:label_fingerprint;type:varchar(255)" json:"labelFingerprint"`
	LoggingService             string    `gorm:"column:logging_service;type:varchar(255)" json:"loggingService"`
	MonitoringService          string    `gorm:"column:monitoring_service;type:varchar(255)" json:"monitoringService"`
	EnableKubernetesAlpha      bool      `gorm:"column:enable_kubernetes_alpha" json:"enableKubernetesAlpha"`
	EnableTpu                  bool      `gorm:"column:enable_tpu" json:"enableTpu"`
	TpuIpv4CidrBlock           string    `gorm:"column:tpu_ipv4_cidr_block;type:varchar(50)" json:"tpuIpv4CidrBlock"`
	AddonsConfigJSON           string    `gorm:"column:addons_config_json;type:jsonb" json:"addonsConfig"`
	PrivateClusterConfigJSON   string    `gorm:"column:private_cluster_config_json;type:jsonb" json:"privateClusterConfig"`
	IpAllocationPolicyJSON     string    `gorm:"column:ip_allocation_policy_json;type:jsonb" json:"ipAllocationPolicy"`
	NetworkConfigJSON          string    `gorm:"column:network_config_json;type:jsonb" json:"networkConfig"`
	MasterAuthJSON             string    `gorm:"column:master_auth_json;type:jsonb" json:"masterAuth"`
	AutoscalingJSON            string    `gorm:"column:autoscaling_json;type:jsonb" json:"autoscaling"`
	VerticalPodAutoscalingJSON string    `gorm:"column:vertical_pod_autoscaling_json;type:jsonb" json:"verticalPodAutoscaling"`
	MonitoringConfigJSON       string    `gorm:"column:monitoring_config_json;type:jsonb" json:"monitoringConfig"`
	LoggingConfigJSON          string    `gorm:"column:logging_config_json;type:jsonb" json:"loggingConfig"`
	MaintenancePolicyJSON      string    `gorm:"column:maintenance_policy_json;type:jsonb" json:"maintenancePolicy"`
	DatabaseEncryptionJSON     string    `gorm:"column:database_encryption_json;type:jsonb" json:"databaseEncryption"`
	WorkloadIdentityConfigJSON string    `gorm:"column:workload_identity_config_json;type:jsonb" json:"workloadIdentityConfig"`
	AutopilotJSON              string    `gorm:"column:autopilot_json;type:jsonb" json:"autopilot"`
	ReleaseChannelJSON         string    `gorm:"column:release_channel_json;type:jsonb" json:"releaseChannel"`
	BinaryAuthorizationJSON    string    `gorm:"column:binary_authorization_json;type:jsonb" json:"binaryAuthorization"`
	SecurityPostureConfigJSON  string    `gorm:"column:security_posture_config_json;type:jsonb" json:"securityPostureConfig"`
	NodePoolDefaultsJSON       string    `gorm:"column:node_pool_defaults_json;type:jsonb" json:"nodePoolDefaults"`
	FleetJSON                  string    `gorm:"column:fleet_json;type:jsonb" json:"fleet"`
	ProjectID                  string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"projectId"`
	CollectedAt                time.Time `gorm:"column:collected_at;not null" json:"collectedAt"`
}

func (GCPContainerCluster) TableName() string {
	return "bronze_history.gcp_container_clusters"
}

// GCPContainerClusterAddon stores historical snapshots of cluster addons.
// Links via ClusterHistoryID, has own valid_from/valid_to for granular tracking.
type GCPContainerClusterAddon struct {
	HistoryID        uint `gorm:"primaryKey"`
	ClusterHistoryID uint `gorm:"column:cluster_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	AddonName  string `gorm:"column:addon_name;type:varchar(100)" json:"addonName"`
	Enabled    bool   `gorm:"column:enabled" json:"enabled"`
	ConfigJSON string `gorm:"column:config_json;type:jsonb" json:"config"`
}

func (GCPContainerClusterAddon) TableName() string {
	return "bronze_history.gcp_container_cluster_addons"
}

// GCPContainerClusterCondition stores historical snapshots of cluster conditions.
// Links via ClusterHistoryID, has own valid_from/valid_to for granular tracking.
type GCPContainerClusterCondition struct {
	HistoryID        uint `gorm:"primaryKey"`
	ClusterHistoryID uint `gorm:"column:cluster_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	Code          string `gorm:"column:code;type:varchar(100)" json:"code"`
	Message       string `gorm:"column:message;type:text" json:"message"`
	CanonicalCode string `gorm:"column:canonical_code;type:varchar(50)" json:"canonicalCode"`
}

func (GCPContainerClusterCondition) TableName() string {
	return "bronze_history.gcp_container_cluster_conditions"
}

// GCPContainerClusterLabel stores historical snapshots of cluster labels.
// Links via ClusterHistoryID, has own valid_from/valid_to for granular tracking.
type GCPContainerClusterLabel struct {
	HistoryID        uint `gorm:"primaryKey"`
	ClusterHistoryID uint `gorm:"column:cluster_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	Key   string `gorm:"column:key;type:varchar(255)" json:"key"`
	Value string `gorm:"column:value;type:varchar(255)" json:"value"`
}

func (GCPContainerClusterLabel) TableName() string {
	return "bronze_history.gcp_container_cluster_labels"
}

// GCPContainerClusterNodePool stores historical snapshots of cluster node pools.
// Links via ClusterHistoryID, has own valid_from/valid_to for granular tracking.
type GCPContainerClusterNodePool struct {
	HistoryID        uint `gorm:"primaryKey"`
	ClusterHistoryID uint `gorm:"column:cluster_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	// All node pool fields (same as bronze.GCPContainerClusterNodePool)
	Name                  string `gorm:"column:name;type:varchar(255)" json:"name"`
	Version               string `gorm:"column:version;type:varchar(50)" json:"version"`
	Status                string `gorm:"column:status;type:varchar(50)" json:"status"`
	StatusMessage         string `gorm:"column:status_message;type:text" json:"statusMessage"`
	InitialNodeCount      int32  `gorm:"column:initial_node_count" json:"initialNodeCount"`
	SelfLink              string `gorm:"column:self_link;type:text" json:"selfLink"`
	PodIpv4CidrSize       int32  `gorm:"column:pod_ipv4_cidr_size" json:"podIpv4CidrSize"`
	Etag                  string `gorm:"column:etag;type:varchar(255)" json:"etag"`
	LocationsJSON         string `gorm:"column:locations_json;type:jsonb" json:"locations"`
	ConfigJSON            string `gorm:"column:config_json;type:jsonb" json:"config"`
	AutoscalingJSON       string `gorm:"column:autoscaling_json;type:jsonb" json:"autoscaling"`
	ManagementJSON        string `gorm:"column:management_json;type:jsonb" json:"management"`
	UpgradeSettingsJSON   string `gorm:"column:upgrade_settings_json;type:jsonb" json:"upgradeSettings"`
	NetworkConfigJSON     string `gorm:"column:network_config_json;type:jsonb" json:"networkConfig"`
	PlacementPolicyJSON   string `gorm:"column:placement_policy_json;type:jsonb" json:"placementPolicy"`
	MaxPodsConstraintJSON string `gorm:"column:max_pods_constraint_json;type:jsonb" json:"maxPodsConstraint"`
}

func (GCPContainerClusterNodePool) TableName() string {
	return "bronze_history.gcp_container_cluster_node_pools"
}
