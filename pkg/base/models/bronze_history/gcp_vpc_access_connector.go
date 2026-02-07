package bronze_history

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPVpcAccessConnector stores historical snapshots of GCP VPC Access connectors.
// Uses ResourceID for lookup (full resource name), with valid_from/valid_to for time range.
type GCPVpcAccessConnector struct {
	HistoryID uint `gorm:"primaryKey"`

	// Link by full resource name
	ResourceID string `gorm:"column:resource_id;type:text;not null;index" json:"name"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"` // null = current

	// All connector fields (same as bronze.GCPVpcAccessConnector)
	Network       string `gorm:"column:network;type:text" json:"network"`
	IpCidrRange   string `gorm:"column:ip_cidr_range;type:varchar(50)" json:"ipCidrRange"`
	State         string `gorm:"column:state;type:varchar(50)" json:"state"`
	MinThroughput int32  `gorm:"column:min_throughput" json:"minThroughput"`
	MaxThroughput int32  `gorm:"column:max_throughput" json:"maxThroughput"`
	MinInstances  int32  `gorm:"column:min_instances" json:"minInstances"`
	MaxInstances  int32  `gorm:"column:max_instances" json:"maxInstances"`
	MachineType   string `gorm:"column:machine_type;type:varchar(255)" json:"machineType"`
	Region        string `gorm:"column:region;type:varchar(255)" json:"region"`

	// JSONB fields
	SubnetJSON            jsonb.JSON `gorm:"column:subnet_json;type:jsonb" json:"subnet"`
	ConnectedProjectsJSON jsonb.JSON `gorm:"column:connected_projects_json;type:jsonb" json:"connectedProjects"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"projectId"`
	CollectedAt time.Time `gorm:"column:collected_at;not null" json:"collectedAt"`
}

func (GCPVpcAccessConnector) TableName() string {
	return "bronze_history.gcp_vpc_access_connectors"
}
