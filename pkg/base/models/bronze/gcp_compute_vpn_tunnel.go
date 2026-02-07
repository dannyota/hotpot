package bronze

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeVpnTunnel represents a GCP Compute Engine VPN tunnel in the bronze layer.
// Fields preserve raw API response data from compute.vpnTunnels.aggregatedList.
type GCPComputeVpnTunnel struct {
	ResourceID        string `gorm:"primaryKey;column:resource_id;type:varchar(255)" json:"id"`
	Name              string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description       string `gorm:"column:description;type:text" json:"description"`
	Status            string `gorm:"column:status;type:varchar(50);index" json:"status"`
	DetailedStatus    string `gorm:"column:detailed_status;type:text" json:"detailedStatus"`
	Region            string `gorm:"column:region;type:text" json:"region"`
	SelfLink          string `gorm:"column:self_link;type:text" json:"selfLink"`
	CreationTimestamp string `gorm:"column:creation_timestamp;type:varchar(50)" json:"creationTimestamp"`
	LabelFingerprint  string `gorm:"column:label_fingerprint;type:varchar(255)" json:"labelFingerprint"`

	// IKE settings
	IkeVersion int32 `gorm:"column:ike_version" json:"ikeVersion"`

	// Peer settings
	PeerIp                       string `gorm:"column:peer_ip;type:varchar(255)" json:"peerIp"`
	PeerExternalGateway          string `gorm:"column:peer_external_gateway;type:text" json:"peerExternalGateway"`
	PeerExternalGatewayInterface int32  `gorm:"column:peer_external_gateway_interface" json:"peerExternalGatewayInterface"`
	PeerGcpGateway               string `gorm:"column:peer_gcp_gateway;type:text" json:"peerGcpGateway"`

	// Routing
	Router string `gorm:"column:router;type:text" json:"router"`

	// Security (SharedSecret excluded - sensitive)
	SharedSecretHash string `gorm:"column:shared_secret_hash;type:varchar(255)" json:"sharedSecretHash"`

	// Gateway references
	VpnGateway          string `gorm:"column:vpn_gateway;type:text" json:"vpnGateway"`
	TargetVpnGateway    string `gorm:"column:target_vpn_gateway;type:text" json:"targetVpnGateway"`
	VpnGatewayInterface int32  `gorm:"column:vpn_gateway_interface" json:"vpnGatewayInterface"`

	// LocalTrafficSelectorJSON contains local CIDR ranges.
	//
	//	["10.0.0.0/8", "192.168.0.0/16"]
	LocalTrafficSelectorJSON jsonb.JSON `gorm:"column:local_traffic_selector_json;type:jsonb" json:"localTrafficSelector"`

	// RemoteTrafficSelectorJSON contains remote CIDR ranges.
	//
	//	["172.16.0.0/12"]
	RemoteTrafficSelectorJSON jsonb.JSON `gorm:"column:remote_traffic_selector_json;type:jsonb" json:"remoteTrafficSelector"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"-"`
	CollectedAt time.Time `gorm:"column:collected_at;not null;index" json:"-"`

	// Relationships
	Labels []GCPComputeVpnTunnelLabel `gorm:"foreignKey:VpnTunnelResourceID;references:ResourceID" json:"labels,omitempty"`
}

func (GCPComputeVpnTunnel) TableName() string {
	return "bronze.gcp_compute_vpn_tunnels"
}

// GCPComputeVpnTunnelLabel represents a label attached to a VPN tunnel.
type GCPComputeVpnTunnelLabel struct {
	ID                  uint   `gorm:"primaryKey"`
	VpnTunnelResourceID string `gorm:"column:vpn_tunnel_resource_id;type:varchar(255);not null;index" json:"-"`
	Key                 string `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value               string `gorm:"column:value;type:text" json:"value"`
}

func (GCPComputeVpnTunnelLabel) TableName() string {
	return "bronze.gcp_compute_vpn_tunnel_labels"
}
