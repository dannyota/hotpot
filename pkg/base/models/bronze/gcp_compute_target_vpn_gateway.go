package bronze

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeTargetVpnGateway represents a GCP Compute Engine Classic VPN gateway in the bronze layer.
// Fields preserve raw API response data from compute.targetVpnGateways.aggregatedList.
type GCPComputeTargetVpnGateway struct {
	ResourceID        string `gorm:"primaryKey;column:resource_id;type:varchar(255)" json:"id"`
	Name              string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description       string `gorm:"column:description;type:text" json:"description"`
	Status            string `gorm:"column:status;type:varchar(50);index" json:"status"`
	Region            string `gorm:"column:region;type:text" json:"region"`
	Network           string `gorm:"column:network;type:text" json:"network"`
	SelfLink          string `gorm:"column:self_link;type:text" json:"selfLink"`
	CreationTimestamp string `gorm:"column:creation_timestamp;type:varchar(50)" json:"creationTimestamp"`
	LabelFingerprint  string `gorm:"column:label_fingerprint;type:varchar(255)" json:"labelFingerprint"`

	// ForwardingRulesJSON contains forwarding rule URLs.
	//
	//	["https://www.googleapis.com/compute/v1/projects/.../forwardingRules/..."]
	ForwardingRulesJSON jsonb.JSON `gorm:"column:forwarding_rules_json;type:jsonb" json:"forwardingRules"`

	// TunnelsJSON contains VPN tunnel URLs.
	//
	//	["https://www.googleapis.com/compute/v1/projects/.../vpnTunnels/..."]
	TunnelsJSON jsonb.JSON `gorm:"column:tunnels_json;type:jsonb" json:"tunnels"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"-"`
	CollectedAt time.Time `gorm:"column:collected_at;not null;index" json:"-"`

	// Relationships
	Labels []GCPComputeTargetVpnGatewayLabel `gorm:"foreignKey:TargetVpnGatewayResourceID;references:ResourceID" json:"labels,omitempty"`
}

func (GCPComputeTargetVpnGateway) TableName() string {
	return "bronze.gcp_compute_target_vpn_gateways"
}

// GCPComputeTargetVpnGatewayLabel represents a label attached to a Classic VPN gateway.
type GCPComputeTargetVpnGatewayLabel struct {
	ID                         uint   `gorm:"primaryKey"`
	TargetVpnGatewayResourceID string `gorm:"column:target_vpn_gateway_resource_id;type:varchar(255);not null;index" json:"-"`
	Key                        string `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value                      string `gorm:"column:value;type:text" json:"value"`
}

func (GCPComputeTargetVpnGatewayLabel) TableName() string {
	return "bronze.gcp_compute_target_vpn_gateway_labels"
}
