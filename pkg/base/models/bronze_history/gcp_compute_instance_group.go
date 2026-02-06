package bronze_history

import (
	"time"
)

// GCPComputeInstanceGroup stores historical snapshots of GCP Compute instance groups.
// Uses ResourceID for lookup (has API ID), with valid_from/valid_to for time range.
type GCPComputeInstanceGroup struct {
	HistoryID uint `gorm:"primaryKey"`

	// Link by API ID (InstanceGroup has unique API ID)
	ResourceID string `gorm:"column:resource_id;type:varchar(255);not null;index" json:"id"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"` // null = current

	// All instance group fields (same as bronze.GCPComputeInstanceGroup)
	Name              string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description       string `gorm:"column:description;type:text" json:"description"`
	Zone              string `gorm:"column:zone;type:text" json:"zone"`
	Network           string `gorm:"column:network;type:text" json:"network"`
	Subnetwork        string `gorm:"column:subnetwork;type:text" json:"subnetwork"`
	Size              int32  `gorm:"column:size" json:"size"`
	SelfLink          string `gorm:"column:self_link;type:text" json:"selfLink"`
	CreationTimestamp string `gorm:"column:creation_timestamp;type:varchar(50)" json:"creationTimestamp"`
	Fingerprint       string `gorm:"column:fingerprint;type:varchar(255)" json:"fingerprint"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"projectId"`
	CollectedAt time.Time `gorm:"column:collected_at;not null" json:"collectedAt"`
}

func (GCPComputeInstanceGroup) TableName() string {
	return "bronze_history.gcp_compute_instance_groups"
}

// GCPComputeInstanceGroupNamedPort stores historical snapshots of instance group named ports.
// Links via GroupHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeInstanceGroupNamedPort struct {
	HistoryID      uint `gorm:"primaryKey"`
	GroupHistoryID uint `gorm:"column:group_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	// Named port fields
	Name string `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Port int32  `gorm:"column:port" json:"port"`
}

func (GCPComputeInstanceGroupNamedPort) TableName() string {
	return "bronze_history.gcp_compute_instance_group_named_ports"
}

// GCPComputeInstanceGroupMember stores historical snapshots of instance group members.
// Links via GroupHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeInstanceGroupMember struct {
	HistoryID      uint `gorm:"primaryKey"`
	GroupHistoryID uint `gorm:"column:group_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	// Member fields
	InstanceURL  string `gorm:"column:instance_url;type:text" json:"instance"`
	InstanceName string `gorm:"column:instance_name;type:varchar(255)" json:"-"`
	Status       string `gorm:"column:status;type:varchar(50)" json:"status"`
}

func (GCPComputeInstanceGroupMember) TableName() string {
	return "bronze_history.gcp_compute_instance_group_members"
}
