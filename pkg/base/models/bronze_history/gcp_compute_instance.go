package bronze_history

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeInstance stores historical snapshots of GCP Compute instances.
// Uses ResourceID for lookup (has API ID), with valid_from/valid_to for time range.
type GCPComputeInstance struct {
	HistoryID uint `gorm:"primaryKey"`

	// Link by API ID (Instance has unique API ID)
	ResourceID string `gorm:"column:resource_id;type:varchar(255);not null;index" json:"id"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"` // null = current

	// All instance fields (same as bronze.GCPComputeInstance)
	Name                   string     `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Zone                   string     `gorm:"column:zone;type:text" json:"zone"`
	MachineType            string     `gorm:"column:machine_type;type:text" json:"machineType"`
	Status                 string     `gorm:"column:status;type:varchar(50)" json:"status"`
	StatusMessage          string     `gorm:"column:status_message;type:text" json:"statusMessage"`
	CpuPlatform            string     `gorm:"column:cpu_platform;type:varchar(100)" json:"cpuPlatform"`
	Hostname               string     `gorm:"column:hostname;type:varchar(255)" json:"hostname"`
	Description            string     `gorm:"column:description;type:text" json:"description"`
	CreationTimestamp      string     `gorm:"column:creation_timestamp;type:varchar(50)" json:"creationTimestamp"`
	LastStartTimestamp     string     `gorm:"column:last_start_timestamp;type:varchar(50)" json:"lastStartTimestamp"`
	LastStopTimestamp      string     `gorm:"column:last_stop_timestamp;type:varchar(50)" json:"lastStopTimestamp"`
	LastSuspendedTimestamp string     `gorm:"column:last_suspended_timestamp;type:varchar(50)" json:"lastSuspendedTimestamp"`
	DeletionProtection     bool       `gorm:"column:deletion_protection" json:"deletionProtection"`
	CanIpForward           bool       `gorm:"column:can_ip_forward" json:"canIpForward"`
	SelfLink               string     `gorm:"column:self_link;type:text" json:"selfLink"`
	SchedulingJSON         jsonb.JSON `gorm:"column:scheduling_json;type:jsonb" json:"scheduling"`
	ProjectID              string     `gorm:"column:project_id;type:varchar(255);not null;index" json:"projectId"`
	CollectedAt            time.Time  `gorm:"column:collected_at;not null" json:"collectedAt"`
}

func (GCPComputeInstance) TableName() string {
	return "bronze_history.gcp_compute_instances"
}

// GCPComputeInstanceDisk stores historical snapshots of instance attached disks.
// Links via InstanceHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeInstanceDisk struct {
	HistoryID         uint `gorm:"primaryKey"`
	InstanceHistoryID uint `gorm:"column:instance_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	// All disk fields (same as bronze.GCPComputeInstanceDisk)
	Source                string     `gorm:"column:source;type:text" json:"source"`
	DeviceName            string     `gorm:"column:device_name;type:varchar(255)" json:"deviceName"`
	Index                 int        `gorm:"column:index" json:"index"`
	Boot                  bool       `gorm:"column:boot" json:"boot"`
	AutoDelete            bool       `gorm:"column:auto_delete" json:"autoDelete"`
	Mode                  string     `gorm:"column:mode;type:varchar(50)" json:"mode"`
	Interface             string     `gorm:"column:interface;type:varchar(50)" json:"interface"`
	Type                  string     `gorm:"column:type;type:varchar(50)" json:"type"`
	DiskSizeGb            int64      `gorm:"column:disk_size_gb" json:"diskSizeGb"`
	DiskEncryptionKeyJSON jsonb.JSON `gorm:"column:disk_encryption_key_json;type:jsonb" json:"diskEncryptionKey"`
	InitializeParamsJSON  jsonb.JSON `gorm:"column:initialize_params_json;type:jsonb" json:"initializeParams"`
}

func (GCPComputeInstanceDisk) TableName() string {
	return "bronze_history.gcp_compute_instance_disks"
}

// GCPComputeInstanceDiskLicense stores historical snapshots of disk licenses.
// Links via DiskHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeInstanceDiskLicense struct {
	HistoryID     uint `gorm:"primaryKey"`
	DiskHistoryID uint `gorm:"column:disk_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	License string `gorm:"column:license;type:text" json:"license"`
}

func (GCPComputeInstanceDiskLicense) TableName() string {
	return "bronze_history.gcp_compute_instance_disk_licenses"
}

// GCPComputeInstanceLabel stores historical snapshots of instance labels.
// Links via InstanceHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeInstanceLabel struct {
	HistoryID         uint `gorm:"primaryKey"`
	InstanceHistoryID uint `gorm:"column:instance_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	Key   string `gorm:"column:key;type:varchar(255)" json:"key"`
	Value string `gorm:"column:value;type:text" json:"value"`
}

func (GCPComputeInstanceLabel) TableName() string {
	return "bronze_history.gcp_compute_instance_labels"
}

// GCPComputeInstanceMetadata stores historical snapshots of instance metadata.
// Links via InstanceHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeInstanceMetadata struct {
	HistoryID         uint `gorm:"primaryKey"`
	InstanceHistoryID uint `gorm:"column:instance_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	Key   string `gorm:"column:key;type:varchar(255)" json:"key"`
	Value string `gorm:"column:value;type:text" json:"value"`
}

func (GCPComputeInstanceMetadata) TableName() string {
	return "bronze_history.gcp_compute_instance_metadata"
}

// GCPComputeInstanceNIC stores historical snapshots of instance network interfaces.
// Links via InstanceHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeInstanceNIC struct {
	HistoryID         uint `gorm:"primaryKey"`
	InstanceHistoryID uint `gorm:"column:instance_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	// All NIC fields (same as bronze.GCPComputeInstanceNIC)
	Name       string `gorm:"column:name;type:varchar(255)" json:"name"`
	Network    string `gorm:"column:network;type:text" json:"network"`
	Subnetwork string `gorm:"column:subnetwork;type:text" json:"subnetwork"`
	NetworkIP  string `gorm:"column:network_ip;type:varchar(50)" json:"networkIP"`
	StackType  string `gorm:"column:stack_type;type:varchar(50)" json:"stackType"`
	NicType    string `gorm:"column:nic_type;type:varchar(50)" json:"nicType"`
}

func (GCPComputeInstanceNIC) TableName() string {
	return "bronze_history.gcp_compute_instance_nics"
}

// GCPComputeInstanceNICAccessConfig stores historical snapshots of NIC access configs.
// Links via NICHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeInstanceNICAccessConfig struct {
	HistoryID    uint `gorm:"primaryKey"`
	NICHistoryID uint `gorm:"column:nic_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	// All access config fields
	Type        string `gorm:"column:type;type:varchar(50)" json:"type"`
	Name        string `gorm:"column:name;type:varchar(255)" json:"name"`
	NatIP       string `gorm:"column:nat_ip;type:varchar(50)" json:"natIP"`
	NetworkTier string `gorm:"column:network_tier;type:varchar(50)" json:"networkTier"`
}

func (GCPComputeInstanceNICAccessConfig) TableName() string {
	return "bronze_history.gcp_compute_instance_nic_access_configs"
}

// GCPComputeInstanceNICAliasRange stores historical snapshots of NIC alias IP ranges.
// Links via NICHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeInstanceNICAliasRange struct {
	HistoryID    uint `gorm:"primaryKey"`
	NICHistoryID uint `gorm:"column:nic_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	IpCidrRange         string `gorm:"column:ip_cidr_range;type:varchar(50)" json:"ipCidrRange"`
	SubnetworkRangeName string `gorm:"column:subnetwork_range_name;type:varchar(255)" json:"subnetworkRangeName"`
}

func (GCPComputeInstanceNICAliasRange) TableName() string {
	return "bronze_history.gcp_compute_instance_nic_alias_ranges"
}

// GCPComputeInstanceServiceAccount stores historical snapshots of instance service accounts.
// Links via InstanceHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeInstanceServiceAccount struct {
	HistoryID         uint `gorm:"primaryKey"`
	InstanceHistoryID uint `gorm:"column:instance_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	Email      string     `gorm:"column:email;type:varchar(255)" json:"email"`
	ScopesJSON jsonb.JSON `gorm:"column:scopes_json;type:jsonb" json:"scopes"`
}

func (GCPComputeInstanceServiceAccount) TableName() string {
	return "bronze_history.gcp_compute_instance_service_accounts"
}

// GCPComputeInstanceTag stores historical snapshots of instance network tags.
// Links via InstanceHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeInstanceTag struct {
	HistoryID         uint `gorm:"primaryKey"`
	InstanceHistoryID uint `gorm:"column:instance_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	Tag string `gorm:"column:tag;type:varchar(255)" json:"tag"`
}

func (GCPComputeInstanceTag) TableName() string {
	return "bronze_history.gcp_compute_instance_tags"
}
