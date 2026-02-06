package bronze

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeInstance represents a GCP Compute Engine instance in the bronze layer.
// Fields preserve raw API response data from compute.instances.list.
type GCPComputeInstance struct {
	// GCP API fields (json tag = original API field name for traceability)
	// ResourceID is the GCP API ID, used as primary key for linking.
	ResourceID             string `gorm:"primaryKey;column:resource_id;type:varchar(255)" json:"id"`
	Name                   string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Zone                   string `gorm:"column:zone;type:text" json:"zone"`
	MachineType            string `gorm:"column:machine_type;type:text" json:"machineType"`
	Status                 string `gorm:"column:status;type:varchar(50);index" json:"status"`
	StatusMessage          string `gorm:"column:status_message;type:text" json:"statusMessage"`
	CpuPlatform            string `gorm:"column:cpu_platform;type:varchar(100)" json:"cpuPlatform"`
	Hostname               string `gorm:"column:hostname;type:varchar(255)" json:"hostname"`
	Description            string `gorm:"column:description;type:text" json:"description"`
	CreationTimestamp      string `gorm:"column:creation_timestamp;type:varchar(50)" json:"creationTimestamp"`
	LastStartTimestamp     string `gorm:"column:last_start_timestamp;type:varchar(50)" json:"lastStartTimestamp"`
	LastStopTimestamp      string `gorm:"column:last_stop_timestamp;type:varchar(50)" json:"lastStopTimestamp"`
	LastSuspendedTimestamp string `gorm:"column:last_suspended_timestamp;type:varchar(50)" json:"lastSuspendedTimestamp"`
	DeletionProtection     bool   `gorm:"column:deletion_protection" json:"deletionProtection"`
	CanIpForward           bool   `gorm:"column:can_ip_forward" json:"canIpForward"`
	SelfLink               string `gorm:"column:self_link;type:text" json:"selfLink"`

	// SchedulingJSON contains VM scheduling configuration.
	//
	//	{
	//	  "preemptible": bool,
	//	  "onHostMaintenance": "MIGRATE" | "TERMINATE",
	//	  "automaticRestart": bool,
	//	  "provisioningModel": "STANDARD" | "SPOT"
	//	}
	SchedulingJSON jsonb.JSON `gorm:"column:scheduling_json;type:jsonb" json:"scheduling"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"-"`
	CollectedAt time.Time `gorm:"column:collected_at;not null;index" json:"-"`

	// Relationships (linked by ResourceID, no FK constraint)
	Disks           []GCPComputeInstanceDisk           `gorm:"foreignKey:InstanceResourceID;references:ResourceID" json:"disks,omitempty"`
	NICs            []GCPComputeInstanceNIC            `gorm:"foreignKey:InstanceResourceID;references:ResourceID" json:"networkInterfaces,omitempty"`
	Labels          []GCPComputeInstanceLabel          `gorm:"foreignKey:InstanceResourceID;references:ResourceID" json:"labels,omitempty"`
	Tags            []GCPComputeInstanceTag            `gorm:"foreignKey:InstanceResourceID;references:ResourceID" json:"tags,omitempty"`
	Metadata        []GCPComputeInstanceMetadata       `gorm:"foreignKey:InstanceResourceID;references:ResourceID" json:"metadata,omitempty"`
	ServiceAccounts []GCPComputeInstanceServiceAccount `gorm:"foreignKey:InstanceResourceID;references:ResourceID" json:"serviceAccounts,omitempty"`
}

func (GCPComputeInstance) TableName() string {
	return "bronze.gcp_compute_instances"
}

// GCPComputeInstanceDisk represents an attached disk on a GCP Compute instance.
// This stores attachment info from instance.disks[], not full disk resource data.
type GCPComputeInstanceDisk struct {
	ID                 uint   `gorm:"primaryKey"`
	InstanceResourceID string `gorm:"column:instance_resource_id;type:varchar(255);not null;index" json:"-"`

	// GCP API fields (json tag = original API field name)
	Source     string `gorm:"column:source;type:text" json:"source"`
	DeviceName string `gorm:"column:device_name;type:varchar(255)" json:"deviceName"`
	Index      int    `gorm:"column:index" json:"index"`
	Boot       bool   `gorm:"column:boot" json:"boot"`
	AutoDelete bool   `gorm:"column:auto_delete" json:"autoDelete"`
	Mode       string `gorm:"column:mode;type:varchar(50)" json:"mode"`
	Interface  string `gorm:"column:interface;type:varchar(50)" json:"interface"`
	Type       string `gorm:"column:type;type:varchar(50)" json:"type"`
	DiskSizeGb int64  `gorm:"column:disk_size_gb" json:"diskSizeGb"`

	// DiskEncryptionKeyJSON contains disk encryption configuration.
	//
	//	{
	//	  "kmsKeyName": "projects/.../cryptoKeys/...",
	//	  "sha256": "base64-encoded-hash"
	//	}
	DiskEncryptionKeyJSON jsonb.JSON `gorm:"column:disk_encryption_key_json;type:jsonb" json:"diskEncryptionKey"`

	// InitializeParamsJSON contains boot disk creation parameters.
	//
	//	{
	//	  "sourceImage": "projects/.../images/...",
	//	  "diskType": "pd-balanced" | "pd-ssd" | "pd-standard",
	//	  "diskSizeGb": "100"
	//	}
	InitializeParamsJSON jsonb.JSON `gorm:"column:initialize_params_json;type:jsonb" json:"initializeParams"`

	// Relationships
	Licenses []GCPComputeInstanceDiskLicense `gorm:"foreignKey:DiskID;constraint:OnDelete:CASCADE" json:"licenses,omitempty"`
}

func (GCPComputeInstanceDisk) TableName() string {
	return "bronze.gcp_compute_instance_disks"
}

// GCPComputeInstanceDiskLicense represents a software license on an attached disk.
// Data from instance.disks[].licenses[].
type GCPComputeInstanceDiskLicense struct {
	ID      uint   `gorm:"primaryKey"`
	DiskID  uint   `gorm:"column:disk_id;not null;index" json:"-"`
	License string `gorm:"column:license;type:text;not null" json:"license"`
}

func (GCPComputeInstanceDiskLicense) TableName() string {
	return "bronze.gcp_compute_instance_disk_licenses"
}

// GCPComputeInstanceLabel represents a label on a GCP Compute instance.
// Data from instance.labels map.
type GCPComputeInstanceLabel struct {
	ID                 uint   `gorm:"primaryKey"`
	InstanceResourceID string `gorm:"column:instance_resource_id;type:varchar(255);not null;index" json:"-"`
	Key                string `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value              string `gorm:"column:value;type:varchar(255)" json:"value"`
}

func (GCPComputeInstanceLabel) TableName() string {
	return "bronze.gcp_compute_instance_labels"
}

// GCPComputeInstanceMetadata represents instance metadata key-value pairs.
// Data from instance.metadata.items[].
type GCPComputeInstanceMetadata struct {
	ID                 uint   `gorm:"primaryKey"`
	InstanceResourceID string `gorm:"column:instance_resource_id;type:varchar(255);not null;index" json:"-"`
	Key                string `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value              string `gorm:"column:value;type:text" json:"value"`
}

func (GCPComputeInstanceMetadata) TableName() string {
	return "bronze.gcp_compute_instance_metadata"
}

// GCPComputeInstanceNIC represents a network interface on a GCP Compute instance.
// Data from instance.networkInterfaces[].
type GCPComputeInstanceNIC struct {
	ID                 uint   `gorm:"primaryKey"`
	InstanceResourceID string `gorm:"column:instance_resource_id;type:varchar(255);not null;index" json:"-"`

	// GCP API fields (json tag = original API field name)
	Name       string `gorm:"column:name;type:varchar(255)" json:"name"`
	Network    string `gorm:"column:network;type:text" json:"network"`
	Subnetwork string `gorm:"column:subnetwork;type:text" json:"subnetwork"`
	NetworkIP  string `gorm:"column:network_ip;type:varchar(50)" json:"networkIP"`
	StackType  string `gorm:"column:stack_type;type:varchar(50)" json:"stackType"`
	NicType    string `gorm:"column:nic_type;type:varchar(50)" json:"nicType"`

	// Relationships
	AccessConfigs []GCPComputeInstanceNICAccessConfig `gorm:"foreignKey:NICID;constraint:OnDelete:CASCADE" json:"accessConfigs,omitempty"`
	AliasIpRanges []GCPComputeInstanceNICAliasRange   `gorm:"foreignKey:NICID;constraint:OnDelete:CASCADE" json:"aliasIpRanges,omitempty"`
}

func (GCPComputeInstanceNIC) TableName() string {
	return "bronze.gcp_compute_instance_nics"
}

// GCPComputeInstanceNICAccessConfig represents external IP configuration for a NIC.
// Data from instance.networkInterfaces[].accessConfigs[].
type GCPComputeInstanceNICAccessConfig struct {
	ID          uint   `gorm:"primaryKey"`
	NICID       uint   `gorm:"column:nic_id;not null;index" json:"-"`
	Type        string `gorm:"column:type;type:varchar(50)" json:"type"`
	Name        string `gorm:"column:name;type:varchar(255)" json:"name"`
	NatIP       string `gorm:"column:nat_ip;type:varchar(50)" json:"natIP"`
	NetworkTier string `gorm:"column:network_tier;type:varchar(50)" json:"networkTier"`
}

func (GCPComputeInstanceNICAccessConfig) TableName() string {
	return "bronze.gcp_compute_instance_nic_access_configs"
}

// GCPComputeInstanceNICAliasRange represents a secondary IP range on a NIC.
// Data from instance.networkInterfaces[].aliasIpRanges[].
type GCPComputeInstanceNICAliasRange struct {
	ID                  uint   `gorm:"primaryKey"`
	NICID               uint   `gorm:"column:nic_id;not null;index" json:"-"`
	IpCidrRange         string `gorm:"column:ip_cidr_range;type:varchar(50)" json:"ipCidrRange"`
	SubnetworkRangeName string `gorm:"column:subnetwork_range_name;type:varchar(255)" json:"subnetworkRangeName"`
}

func (GCPComputeInstanceNICAliasRange) TableName() string {
	return "bronze.gcp_compute_instance_nic_alias_ranges"
}

// GCPComputeInstanceServiceAccount represents a service account attached to an instance.
// Data from instance.serviceAccounts[].
type GCPComputeInstanceServiceAccount struct {
	ID                 uint   `gorm:"primaryKey"`
	InstanceResourceID string `gorm:"column:instance_resource_id;type:varchar(255);not null;index" json:"-"`
	Email              string `gorm:"column:email;type:varchar(255);not null" json:"email"`

	// ScopesJSON contains OAuth scopes granted to the service account.
	//
	//	["https://www.googleapis.com/auth/cloud-platform", ...]
	ScopesJSON jsonb.JSON `gorm:"column:scopes_json;type:jsonb" json:"scopes"`
}

func (GCPComputeInstanceServiceAccount) TableName() string {
	return "bronze.gcp_compute_instance_service_accounts"
}

// GCPComputeInstanceTag represents a network tag on a GCP Compute instance.
// Data from instance.tags.items[].
type GCPComputeInstanceTag struct {
	ID                 uint   `gorm:"primaryKey"`
	InstanceResourceID string `gorm:"column:instance_resource_id;type:varchar(255);not null;index" json:"-"`
	Tag                string `gorm:"column:tag;type:varchar(255);not null" json:"tag"`
}

func (GCPComputeInstanceTag) TableName() string {
	return "bronze.gcp_compute_instance_tags"
}
