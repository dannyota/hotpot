package bronze

import (
	"time"
)

// GCPComputeInstanceGroup represents a GCP Compute Engine instance group in the bronze layer.
// Fields preserve raw API response data from compute.instanceGroups.aggregatedList.
type GCPComputeInstanceGroup struct {
	// GCP API fields (json tag = original API field name for traceability)
	// ResourceID is the GCP API ID, used as primary key for linking.
	ResourceID        string `gorm:"primaryKey;column:resource_id;type:varchar(255)" json:"id"`
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
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"-"`
	CollectedAt time.Time `gorm:"column:collected_at;not null;index" json:"-"`

	// Relationships (linked by ResourceID, no FK constraint)
	NamedPorts []GCPComputeInstanceGroupNamedPort `gorm:"foreignKey:GroupResourceID;references:ResourceID" json:"namedPorts,omitempty"`
	Members    []GCPComputeInstanceGroupMember    `gorm:"foreignKey:GroupResourceID;references:ResourceID" json:"members,omitempty"`
}

func (GCPComputeInstanceGroup) TableName() string {
	return "bronze.gcp_compute_instance_groups"
}

// GCPComputeInstanceGroupNamedPort represents a named port on a GCP instance group.
type GCPComputeInstanceGroupNamedPort struct {
	ID              uint   `gorm:"primaryKey"`
	GroupResourceID string `gorm:"column:group_resource_id;type:varchar(255);not null;index" json:"-"`
	Name            string `gorm:"column:name;type:varchar(100);not null" json:"name"`
	Port            int32  `gorm:"column:port" json:"port"`
}

func (GCPComputeInstanceGroupNamedPort) TableName() string {
	return "bronze.gcp_compute_instance_group_named_ports"
}

// GCPComputeInstanceGroupMember represents a member instance in a GCP instance group.
type GCPComputeInstanceGroupMember struct {
	ID              uint   `gorm:"primaryKey"`
	GroupResourceID string `gorm:"column:group_resource_id;type:varchar(255);not null;index" json:"-"`
	InstanceURL     string `gorm:"column:instance_url;type:text" json:"instance"`
	InstanceName    string `gorm:"column:instance_name;type:varchar(255)" json:"-"`
	Status          string `gorm:"column:status;type:varchar(50)" json:"status"`
}

func (GCPComputeInstanceGroupMember) TableName() string {
	return "bronze.gcp_compute_instance_group_members"
}
