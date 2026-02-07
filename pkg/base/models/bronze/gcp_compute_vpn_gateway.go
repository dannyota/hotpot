package bronze

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeVpnGateway represents a GCP Compute Engine VPN gateway (HA) in the bronze layer.
// Fields preserve raw API response data from compute.vpnGateways.aggregatedList.
type GCPComputeVpnGateway struct {
	// GCP API fields (json tag = original API field name for traceability)
	// ResourceID is the GCP API ID, used as primary key for linking.
	ResourceID        string `gorm:"primaryKey;column:resource_id;type:varchar(255)" json:"id"`
	Name              string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description       string `gorm:"column:description;type:text" json:"description"`
	Region            string `gorm:"column:region;type:text" json:"region"`
	Network           string `gorm:"column:network;type:text" json:"network"`
	SelfLink          string `gorm:"column:self_link;type:text" json:"selfLink"`
	CreationTimestamp string `gorm:"column:creation_timestamp;type:varchar(50)" json:"creationTimestamp"`
	LabelFingerprint  string `gorm:"column:label_fingerprint;type:varchar(255)" json:"labelFingerprint"`
	GatewayIpVersion  string `gorm:"column:gateway_ip_version;type:varchar(50)" json:"gatewayIpVersion"`
	StackType         string `gorm:"column:stack_type;type:varchar(50)" json:"stackType"`

	// VpnInterfacesJSON contains the VPN gateway interfaces configuration.
	//
	//	[{"id": 0, "ipAddress": "1.2.3.4"}, {"id": 1, "ipAddress": "5.6.7.8"}]
	VpnInterfacesJSON jsonb.JSON `gorm:"column:vpn_interfaces_json;type:jsonb" json:"vpnInterfaces"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"-"`
	CollectedAt time.Time `gorm:"column:collected_at;not null;index" json:"-"`

	// Relationships (linked by ResourceID, no FK constraint)
	Labels []GCPComputeVpnGatewayLabel `gorm:"foreignKey:VpnGatewayResourceID;references:ResourceID" json:"labels,omitempty"`
}

func (GCPComputeVpnGateway) TableName() string {
	return "bronze.gcp_compute_vpn_gateways"
}

// GCPComputeVpnGatewayLabel represents a label attached to a GCP Compute VPN gateway.
type GCPComputeVpnGatewayLabel struct {
	ID                    uint   `gorm:"primaryKey"`
	VpnGatewayResourceID string `gorm:"column:vpn_gateway_resource_id;type:varchar(255);not null;index" json:"-"`
	Key                   string `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value                 string `gorm:"column:value;type:text" json:"value"`
}

func (GCPComputeVpnGatewayLabel) TableName() string {
	return "bronze.gcp_compute_vpn_gateway_labels"
}
