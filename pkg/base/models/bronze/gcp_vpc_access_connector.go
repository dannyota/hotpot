package bronze

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPVpcAccessConnector represents a GCP Serverless VPC Access connector in the bronze layer.
// Fields preserve raw API response data from vpcaccess.connectors.list.
type GCPVpcAccessConnector struct {
	// ResourceID is the full resource name (projects/{project}/locations/{region}/connectors/{name}).
	ResourceID    string `gorm:"primaryKey;column:resource_id;type:text" json:"name"`
	Network       string `gorm:"column:network;type:text" json:"network"`
	IpCidrRange   string `gorm:"column:ip_cidr_range;type:varchar(50)" json:"ipCidrRange"`
	State         string `gorm:"column:state;type:varchar(50)" json:"state"`
	MinThroughput int32  `gorm:"column:min_throughput" json:"minThroughput"`
	MaxThroughput int32  `gorm:"column:max_throughput" json:"maxThroughput"`
	MinInstances  int32  `gorm:"column:min_instances" json:"minInstances"`
	MaxInstances  int32  `gorm:"column:max_instances" json:"maxInstances"`
	MachineType   string `gorm:"column:machine_type;type:varchar(255)" json:"machineType"`
	Region        string `gorm:"column:region;type:varchar(255)" json:"region"`

	// SubnetJSON contains subnet configuration.
	//
	//	{"name": "subnet-name", "projectId": "project-id"}
	SubnetJSON jsonb.JSON `gorm:"column:subnet_json;type:jsonb" json:"subnet"`

	// ConnectedProjectsJSON contains list of projects connected to this connector.
	//
	//	["project-1", "project-2"]
	ConnectedProjectsJSON jsonb.JSON `gorm:"column:connected_projects_json;type:jsonb" json:"connectedProjects"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"-"`
	CollectedAt time.Time `gorm:"column:collected_at;not null;index" json:"-"`
}

func (GCPVpcAccessConnector) TableName() string {
	return "bronze.gcp_vpc_access_connectors"
}
