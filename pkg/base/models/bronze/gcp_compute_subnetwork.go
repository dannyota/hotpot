package bronze

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeSubnetwork represents a GCP VPC subnetwork in the bronze layer.
// Fields preserve raw API response data from compute.subnetworks.aggregatedList.
type GCPComputeSubnetwork struct {
	// GCP API fields (json tag = original API field name for traceability)
	// ResourceID is the GCP API ID, used as primary key for linking.
	ResourceID        string `gorm:"primaryKey;column:resource_id;type:varchar(255)" json:"id"`
	Name              string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description       string `gorm:"column:description;type:text" json:"description"`
	SelfLink          string `gorm:"column:self_link;type:text" json:"selfLink"`
	CreationTimestamp string `gorm:"column:creation_timestamp;type:varchar(50)" json:"creationTimestamp"`

	// Network relationship (URL to parent network)
	Network string `gorm:"column:network;type:text;not null" json:"network"`
	Region  string `gorm:"column:region;type:text;not null" json:"region"`

	// IP configuration
	IpCidrRange    string `gorm:"column:ip_cidr_range;type:varchar(50);not null" json:"ipCidrRange"`
	GatewayAddress string `gorm:"column:gateway_address;type:varchar(50)" json:"gatewayAddress"`

	// Purpose and role
	Purpose string `gorm:"column:purpose;type:varchar(50)" json:"purpose"`
	Role    string `gorm:"column:role;type:varchar(50)" json:"role"`

	// Private Google Access
	PrivateIpGoogleAccess   bool   `gorm:"column:private_ip_google_access" json:"privateIpGoogleAccess"`
	PrivateIpv6GoogleAccess string `gorm:"column:private_ipv6_google_access;type:varchar(50)" json:"privateIpv6GoogleAccess"`

	// Stack type and IPv6
	StackType          string `gorm:"column:stack_type;type:varchar(50)" json:"stackType"`
	Ipv6AccessType     string `gorm:"column:ipv6_access_type;type:varchar(50)" json:"ipv6AccessType"`
	InternalIpv6Prefix string `gorm:"column:internal_ipv6_prefix;type:varchar(50)" json:"internalIpv6Prefix"`
	ExternalIpv6Prefix string `gorm:"column:external_ipv6_prefix;type:varchar(50)" json:"externalIpv6Prefix"`

	// LogConfigJSON contains VPC flow logging configuration.
	//
	//	{
	//	  "enable": true,
	//	  "aggregationInterval": "INTERVAL_5_SEC",
	//	  "flowSampling": 0.5,
	//	  "metadata": "INCLUDE_ALL_METADATA"
	//	}
	LogConfigJSON jsonb.JSON `gorm:"column:log_config_json;type:jsonb" json:"logConfig"`

	// Fingerprint for optimistic locking
	Fingerprint string `gorm:"column:fingerprint;type:varchar(255)" json:"fingerprint"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"-"`
	CollectedAt time.Time `gorm:"column:collected_at;not null;index" json:"-"`

	// Relationships
	SecondaryIpRanges []GCPComputeSubnetworkSecondaryRange `gorm:"foreignKey:SubnetworkResourceID;references:ResourceID" json:"secondaryIpRanges,omitempty"`
}

func (GCPComputeSubnetwork) TableName() string {
	return "bronze.gcp_compute_subnetworks"
}

// GCPComputeSubnetworkSecondaryRange represents a secondary IP range on a subnetwork.
// Data from subnetwork.secondaryIpRanges[].
type GCPComputeSubnetworkSecondaryRange struct {
	ID                   uint   `gorm:"primaryKey"`
	SubnetworkResourceID string `gorm:"column:subnetwork_resource_id;type:varchar(255);not null;index" json:"-"`

	// GCP API fields (json tag = original API field name)
	RangeName   string `gorm:"column:range_name;type:varchar(255);not null" json:"rangeName"`
	IpCidrRange string `gorm:"column:ip_cidr_range;type:varchar(50);not null" json:"ipCidrRange"`
}

func (GCPComputeSubnetworkSecondaryRange) TableName() string {
	return "bronze.gcp_compute_subnetwork_secondary_ranges"
}
