package bronze_history

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeAddress stores historical snapshots of GCP Compute regional addresses.
// Uses ResourceID for lookup (has API ID), with valid_from/valid_to for time range.
type GCPComputeAddress struct {
	HistoryID uint `gorm:"primaryKey"`

	// Link by API ID (Address has unique API ID)
	ResourceID string `gorm:"column:resource_id;type:varchar(255);not null;index" json:"id"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"` // null = current

	// All address fields (same as bronze.GCPComputeAddress)
	Name              string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description       string `gorm:"column:description;type:text" json:"description"`
	Address           string `gorm:"column:address;type:varchar(50)" json:"address"`
	AddressType       string `gorm:"column:address_type;type:varchar(50)" json:"addressType"`
	IpVersion         string `gorm:"column:ip_version;type:varchar(10)" json:"ipVersion"`
	Ipv6EndpointType  string `gorm:"column:ipv6_endpoint_type;type:varchar(50)" json:"ipv6EndpointType"`
	IpCollection      string `gorm:"column:ip_collection;type:text" json:"ipCollection"`
	Region            string `gorm:"column:region;type:text" json:"region"`
	Status            string `gorm:"column:status;type:varchar(50)" json:"status"`
	Purpose           string `gorm:"column:purpose;type:varchar(100)" json:"purpose"`
	Network           string `gorm:"column:network;type:text" json:"network"`
	Subnetwork        string `gorm:"column:subnetwork;type:text" json:"subnetwork"`
	NetworkTier       string `gorm:"column:network_tier;type:varchar(50)" json:"networkTier"`
	PrefixLength      int32  `gorm:"column:prefix_length" json:"prefixLength"`
	SelfLink          string `gorm:"column:self_link;type:text" json:"selfLink"`
	CreationTimestamp string `gorm:"column:creation_timestamp;type:varchar(50)" json:"creationTimestamp"`
	LabelFingerprint  string `gorm:"column:label_fingerprint;type:varchar(255)" json:"labelFingerprint"`

	// JSON arrays
	UsersJSON jsonb.JSON `gorm:"column:users_json;type:jsonb" json:"users"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"projectId"`
	CollectedAt time.Time `gorm:"column:collected_at;not null" json:"collectedAt"`
}

func (GCPComputeAddress) TableName() string {
	return "bronze_history.gcp_compute_addresses"
}

// GCPComputeAddressLabel stores historical snapshots of address labels.
// Links via AddressHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeAddressLabel struct {
	HistoryID        uint `gorm:"primaryKey"`
	AddressHistoryID uint `gorm:"column:address_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	// Label fields
	Key   string `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value string `gorm:"column:value;type:text" json:"value"`
}

func (GCPComputeAddressLabel) TableName() string {
	return "bronze_history.gcp_compute_address_labels"
}
