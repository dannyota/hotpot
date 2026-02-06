package bronze_history

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeSubnetwork stores historical snapshots of GCP VPC subnetworks.
// Uses ResourceID for lookup (has API ID), with valid_from/valid_to for time range.
type GCPComputeSubnetwork struct {
	HistoryID uint `gorm:"primaryKey"`

	// Link by API ID (Subnetwork has unique API ID)
	ResourceID string `gorm:"column:resource_id;type:varchar(255);not null;index" json:"id"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"` // null = current

	// All subnetwork fields (same as bronze.GCPComputeSubnetwork)
	Name              string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description       string `gorm:"column:description;type:text" json:"description"`
	SelfLink          string `gorm:"column:self_link;type:text" json:"selfLink"`
	CreationTimestamp string `gorm:"column:creation_timestamp;type:varchar(50)" json:"creationTimestamp"`

	// Network relationship
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

	// LogConfigJSON
	LogConfigJSON jsonb.JSON `gorm:"column:log_config_json;type:jsonb" json:"logConfig"`

	// Fingerprint
	Fingerprint string `gorm:"column:fingerprint;type:varchar(255)" json:"fingerprint"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"projectId"`
	CollectedAt time.Time `gorm:"column:collected_at;not null" json:"collectedAt"`
}

func (GCPComputeSubnetwork) TableName() string {
	return "bronze_history.gcp_compute_subnetworks"
}

// GCPComputeSubnetworkSecondaryRange stores historical snapshots of secondary IP ranges.
// Links via SubnetworkHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeSubnetworkSecondaryRange struct {
	HistoryID           uint `gorm:"primaryKey"`
	SubnetworkHistoryID uint `gorm:"column:subnetwork_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	// All secondary range fields
	RangeName   string `gorm:"column:range_name;type:varchar(255);not null" json:"rangeName"`
	IpCidrRange string `gorm:"column:ip_cidr_range;type:varchar(50);not null" json:"ipCidrRange"`
}

func (GCPComputeSubnetworkSecondaryRange) TableName() string {
	return "bronze_history.gcp_compute_subnetwork_secondary_ranges"
}
