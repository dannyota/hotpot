package bronze

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeAddress represents a GCP Compute Engine regional address in the bronze layer.
// Fields preserve raw API response data from compute.addresses.aggregatedList.
type GCPComputeAddress struct {
	// GCP API fields (json tag = original API field name for traceability)
	// ResourceID is the GCP API ID, used as primary key for linking.
	ResourceID        string `gorm:"primaryKey;column:resource_id;type:varchar(255)" json:"id"`
	Name              string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description       string `gorm:"column:description;type:text" json:"description"`
	Address           string `gorm:"column:address;type:varchar(50)" json:"address"`
	AddressType       string `gorm:"column:address_type;type:varchar(50)" json:"addressType"`
	IpVersion         string `gorm:"column:ip_version;type:varchar(10)" json:"ipVersion"`
	Ipv6EndpointType  string `gorm:"column:ipv6_endpoint_type;type:varchar(50)" json:"ipv6EndpointType"`
	IpCollection      string `gorm:"column:ip_collection;type:text" json:"ipCollection"`
	Region            string `gorm:"column:region;type:text" json:"region"`
	Status            string `gorm:"column:status;type:varchar(50);index" json:"status"`
	Purpose           string `gorm:"column:purpose;type:varchar(100)" json:"purpose"`
	Network           string `gorm:"column:network;type:text" json:"network"`
	Subnetwork        string `gorm:"column:subnetwork;type:text" json:"subnetwork"`
	NetworkTier       string `gorm:"column:network_tier;type:varchar(50)" json:"networkTier"`
	PrefixLength      int32  `gorm:"column:prefix_length" json:"prefixLength"`
	SelfLink          string `gorm:"column:self_link;type:text" json:"selfLink"`
	CreationTimestamp string `gorm:"column:creation_timestamp;type:varchar(50)" json:"creationTimestamp"`
	LabelFingerprint  string `gorm:"column:label_fingerprint;type:varchar(255)" json:"labelFingerprint"`

	// UsersJSON contains list of resource URLs using this address.
	//
	//	["projects/.../instances/vm1", "projects/.../forwardingRules/rule1"]
	UsersJSON jsonb.JSON `gorm:"column:users_json;type:jsonb" json:"users"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"-"`
	CollectedAt time.Time `gorm:"column:collected_at;not null;index" json:"-"`

	// Relationships (linked by ResourceID, no FK constraint)
	Labels []GCPComputeAddressLabel `gorm:"foreignKey:AddressResourceID;references:ResourceID" json:"labels,omitempty"`
}

func (GCPComputeAddress) TableName() string {
	return "bronze.gcp_compute_addresses"
}

// GCPComputeAddressLabel represents a label attached to a GCP Compute address.
type GCPComputeAddressLabel struct {
	ID                uint   `gorm:"primaryKey"`
	AddressResourceID string `gorm:"column:address_resource_id;type:varchar(255);not null;index" json:"-"`
	Key               string `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value             string `gorm:"column:value;type:text" json:"value"`
}

func (GCPComputeAddressLabel) TableName() string {
	return "bronze.gcp_compute_address_labels"
}
