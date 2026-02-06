package bronze

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPContainerCluster represents a GKE cluster in the bronze layer.
// Fields preserve raw API response data from container.clusters.list.
type GCPContainerCluster struct {
	// GCP API fields (json tag = original API field name for traceability)
	ResourceID            string `gorm:"primaryKey;column:resource_id;type:varchar(255)" json:"id"`
	Name                  string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Location              string `gorm:"column:location;type:varchar(100)" json:"location"`
	Zone                  string `gorm:"column:zone;type:varchar(100)" json:"zone"`
	Description           string `gorm:"column:description;type:text" json:"description"`
	InitialClusterVersion string `gorm:"column:initial_cluster_version;type:varchar(50)" json:"initialClusterVersion"`
	CurrentMasterVersion  string `gorm:"column:current_master_version;type:varchar(50)" json:"currentMasterVersion"`
	CurrentNodeVersion    string `gorm:"column:current_node_version;type:varchar(50)" json:"currentNodeVersion"`
	Status                string `gorm:"column:status;type:varchar(50);index" json:"status"`
	StatusMessage         string `gorm:"column:status_message;type:text" json:"statusMessage"`
	CurrentNodeCount      int32  `gorm:"column:current_node_count" json:"currentNodeCount"`
	Network               string `gorm:"column:network;type:varchar(255)" json:"network"`
	Subnetwork            string `gorm:"column:subnetwork;type:varchar(255)" json:"subnetwork"`
	ClusterIpv4Cidr       string `gorm:"column:cluster_ipv4_cidr;type:varchar(50)" json:"clusterIpv4Cidr"`
	ServicesIpv4Cidr      string `gorm:"column:services_ipv4_cidr;type:varchar(50)" json:"servicesIpv4Cidr"`
	NodeIpv4CidrSize      int32  `gorm:"column:node_ipv4_cidr_size" json:"nodeIpv4CidrSize"`
	Endpoint              string `gorm:"column:endpoint;type:varchar(255)" json:"endpoint"`
	SelfLink              string `gorm:"column:self_link;type:text" json:"selfLink"`
	CreateTime            string `gorm:"column:create_time;type:varchar(50)" json:"createTime"`
	ExpireTime            string `gorm:"column:expire_time;type:varchar(50)" json:"expireTime"`
	Etag                  string `gorm:"column:etag;type:varchar(255)" json:"etag"`
	LabelFingerprint      string `gorm:"column:label_fingerprint;type:varchar(255)" json:"labelFingerprint"`
	LoggingService        string `gorm:"column:logging_service;type:varchar(255)" json:"loggingService"`
	MonitoringService     string `gorm:"column:monitoring_service;type:varchar(255)" json:"monitoringService"`
	EnableKubernetesAlpha bool   `gorm:"column:enable_kubernetes_alpha" json:"enableKubernetesAlpha"`
	EnableTpu             bool   `gorm:"column:enable_tpu" json:"enableTpu"`
	TpuIpv4CidrBlock      string `gorm:"column:tpu_ipv4_cidr_block;type:varchar(50)" json:"tpuIpv4CidrBlock"`

	// Nested objects stored as JSONB
	AddonsConfigJSON           jsonb.JSON `gorm:"column:addons_config_json;type:jsonb" json:"addonsConfig"`
	PrivateClusterConfigJSON   jsonb.JSON `gorm:"column:private_cluster_config_json;type:jsonb" json:"privateClusterConfig"`
	IpAllocationPolicyJSON     jsonb.JSON `gorm:"column:ip_allocation_policy_json;type:jsonb" json:"ipAllocationPolicy"`
	NetworkConfigJSON          jsonb.JSON `gorm:"column:network_config_json;type:jsonb" json:"networkConfig"`
	MasterAuthJSON             jsonb.JSON `gorm:"column:master_auth_json;type:jsonb" json:"masterAuth"`
	AutoscalingJSON            jsonb.JSON `gorm:"column:autoscaling_json;type:jsonb" json:"autoscaling"`
	VerticalPodAutoscalingJSON jsonb.JSON `gorm:"column:vertical_pod_autoscaling_json;type:jsonb" json:"verticalPodAutoscaling"`
	MonitoringConfigJSON       jsonb.JSON `gorm:"column:monitoring_config_json;type:jsonb" json:"monitoringConfig"`
	LoggingConfigJSON          jsonb.JSON `gorm:"column:logging_config_json;type:jsonb" json:"loggingConfig"`
	MaintenancePolicyJSON      jsonb.JSON `gorm:"column:maintenance_policy_json;type:jsonb" json:"maintenancePolicy"`
	DatabaseEncryptionJSON     jsonb.JSON `gorm:"column:database_encryption_json;type:jsonb" json:"databaseEncryption"`
	WorkloadIdentityConfigJSON jsonb.JSON `gorm:"column:workload_identity_config_json;type:jsonb" json:"workloadIdentityConfig"`
	AutopilotJSON              jsonb.JSON `gorm:"column:autopilot_json;type:jsonb" json:"autopilot"`
	ReleaseChannelJSON         jsonb.JSON `gorm:"column:release_channel_json;type:jsonb" json:"releaseChannel"`
	BinaryAuthorizationJSON    jsonb.JSON `gorm:"column:binary_authorization_json;type:jsonb" json:"binaryAuthorization"`
	SecurityPostureConfigJSON  jsonb.JSON `gorm:"column:security_posture_config_json;type:jsonb" json:"securityPostureConfig"`
	NodePoolDefaultsJSON       jsonb.JSON `gorm:"column:node_pool_defaults_json;type:jsonb" json:"nodePoolDefaults"`
	FleetJSON                  jsonb.JSON `gorm:"column:fleet_json;type:jsonb" json:"fleet"`

	// Collection metadata (not from API)
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"-"`
	CollectedAt time.Time `gorm:"column:collected_at;not null;index" json:"-"`

	// Relationships
	Labels     []GCPContainerClusterLabel     `gorm:"foreignKey:ClusterResourceID;references:ResourceID" json:"resourceLabels,omitempty"`
	Addons     []GCPContainerClusterAddon     `gorm:"foreignKey:ClusterResourceID;references:ResourceID" json:"-"`
	Conditions []GCPContainerClusterCondition `gorm:"foreignKey:ClusterResourceID;references:ResourceID" json:"conditions,omitempty"`
	NodePools  []GCPContainerClusterNodePool  `gorm:"foreignKey:ClusterResourceID;references:ResourceID" json:"nodePools,omitempty"`
}

func (GCPContainerCluster) TableName() string {
	return "bronze.gcp_container_clusters"
}

// GCPContainerClusterAddon represents an addon configuration for a cluster.
// Data from cluster.addonsConfig, one row per addon.
type GCPContainerClusterAddon struct {
	ID                uint       `gorm:"primaryKey"`
	ClusterResourceID string     `gorm:"column:cluster_resource_id;type:varchar(255);not null;index" json:"-"`
	AddonName         string     `gorm:"column:addon_name;type:varchar(100);not null" json:"addonName"`
	Enabled           bool       `gorm:"column:enabled" json:"enabled"`
	ConfigJSON        jsonb.JSON `gorm:"column:config_json;type:jsonb" json:"config"`
}

func (GCPContainerClusterAddon) TableName() string {
	return "bronze.gcp_container_cluster_addons"
}

// GCPContainerClusterCondition represents a status condition on a cluster.
// Data from cluster.conditions[].
type GCPContainerClusterCondition struct {
	ID                uint   `gorm:"primaryKey"`
	ClusterResourceID string `gorm:"column:cluster_resource_id;type:varchar(255);not null;index" json:"-"`
	Code              string `gorm:"column:code;type:varchar(100)" json:"code"`
	Message           string `gorm:"column:message;type:text" json:"message"`
	CanonicalCode     string `gorm:"column:canonical_code;type:varchar(50)" json:"canonicalCode"`
}

func (GCPContainerClusterCondition) TableName() string {
	return "bronze.gcp_container_cluster_conditions"
}

// GCPContainerClusterLabel represents a resource label on a cluster.
// Data from cluster.resourceLabels map.
type GCPContainerClusterLabel struct {
	ID                uint   `gorm:"primaryKey"`
	ClusterResourceID string `gorm:"column:cluster_resource_id;type:varchar(255);not null;index" json:"-"`
	Key               string `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value             string `gorm:"column:value;type:varchar(255)" json:"value"`
}

func (GCPContainerClusterLabel) TableName() string {
	return "bronze.gcp_container_cluster_labels"
}

// GCPContainerClusterNodePool represents a node pool in a cluster.
// Data from cluster.nodePools[].
type GCPContainerClusterNodePool struct {
	ID                uint   `gorm:"primaryKey"`
	ClusterResourceID string `gorm:"column:cluster_resource_id;type:varchar(255);not null;index" json:"-"`
	Name              string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Version           string `gorm:"column:version;type:varchar(50)" json:"version"`
	Status            string `gorm:"column:status;type:varchar(50)" json:"status"`
	StatusMessage     string `gorm:"column:status_message;type:text" json:"statusMessage"`
	InitialNodeCount  int32  `gorm:"column:initial_node_count" json:"initialNodeCount"`
	SelfLink          string `gorm:"column:self_link;type:text" json:"selfLink"`
	PodIpv4CidrSize   int32  `gorm:"column:pod_ipv4_cidr_size" json:"podIpv4CidrSize"`
	Etag              string `gorm:"column:etag;type:varchar(255)" json:"etag"`

	// Nested objects stored as JSONB (includes locations, config with taints/labels)
	LocationsJSON         jsonb.JSON `gorm:"column:locations_json;type:jsonb" json:"locations"`
	ConfigJSON            jsonb.JSON `gorm:"column:config_json;type:jsonb" json:"config"`
	AutoscalingJSON       jsonb.JSON `gorm:"column:autoscaling_json;type:jsonb" json:"autoscaling"`
	ManagementJSON        jsonb.JSON `gorm:"column:management_json;type:jsonb" json:"management"`
	UpgradeSettingsJSON   jsonb.JSON `gorm:"column:upgrade_settings_json;type:jsonb" json:"upgradeSettings"`
	NetworkConfigJSON     jsonb.JSON `gorm:"column:network_config_json;type:jsonb" json:"networkConfig"`
	PlacementPolicyJSON   jsonb.JSON `gorm:"column:placement_policy_json;type:jsonb" json:"placementPolicy"`
	MaxPodsConstraintJSON jsonb.JSON `gorm:"column:max_pods_constraint_json;type:jsonb" json:"maxPodsConstraint"`
}

func (GCPContainerClusterNodePool) TableName() string {
	return "bronze.gcp_container_cluster_node_pools"
}
