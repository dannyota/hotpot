package bronze_history

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeVpnGateway stores historical snapshots of GCP Compute Engine VPN gateways.
// Uses ResourceID for lookup (has API ID), with valid_from/valid_to for time range.
type GCPComputeVpnGateway struct {
	HistoryID uint `gorm:"primaryKey"`

	// Link by API ID (VPN gateway has unique API ID)
	ResourceID string `gorm:"column:resource_id;type:varchar(255);not null;index" json:"id"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"` // null = current

	// All VPN gateway fields (same as bronze.GCPComputeVpnGateway)
	Name              string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description       string `gorm:"column:description;type:text" json:"description"`
	Region            string `gorm:"column:region;type:text" json:"region"`
	Network           string `gorm:"column:network;type:text" json:"network"`
	SelfLink          string `gorm:"column:self_link;type:text" json:"selfLink"`
	CreationTimestamp string `gorm:"column:creation_timestamp;type:varchar(50)" json:"creationTimestamp"`
	LabelFingerprint  string `gorm:"column:label_fingerprint;type:varchar(255)" json:"labelFingerprint"`
	GatewayIpVersion  string `gorm:"column:gateway_ip_version;type:varchar(50)" json:"gatewayIpVersion"`
	StackType         string `gorm:"column:stack_type;type:varchar(50)" json:"stackType"`

	// JSONB fields
	VpnInterfacesJSON jsonb.JSON `gorm:"column:vpn_interfaces_json;type:jsonb" json:"vpnInterfaces"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"projectId"`
	CollectedAt time.Time `gorm:"column:collected_at;not null" json:"collectedAt"`
}

func (GCPComputeVpnGateway) TableName() string {
	return "bronze_history.gcp_compute_vpn_gateways"
}

// GCPComputeVpnGatewayLabel stores historical snapshots of VPN gateway labels.
// Links via VpnGatewayHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeVpnGatewayLabel struct {
	HistoryID          uint `gorm:"primaryKey"`
	VpnGatewayHistoryID uint `gorm:"column:vpn_gateway_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	// Label fields
	Key   string `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value string `gorm:"column:value;type:text" json:"value"`
}

func (GCPComputeVpnGatewayLabel) TableName() string {
	return "bronze_history.gcp_compute_vpn_gateway_labels"
}
