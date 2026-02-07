package bronze_history

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeVpnTunnel stores historical versions of VPN tunnels.
// Uses ResourceID for lookup (has API ID), with valid_from/valid_to for time range.
type GCPComputeVpnTunnel struct {
	HistoryID uint `gorm:"primaryKey"`

	// Link by API ID (VPN tunnel has unique API ID)
	ResourceID string `gorm:"column:resource_id;type:varchar(255);not null;index" json:"id"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"` // null = current

	// All VPN tunnel fields (same as bronze.GCPComputeVpnTunnel)
	Name              string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description       string `gorm:"column:description;type:text" json:"description"`
	Status            string `gorm:"column:status;type:varchar(50)" json:"status"`
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

	// Security
	SharedSecretHash string `gorm:"column:shared_secret_hash;type:varchar(255)" json:"sharedSecretHash"`

	// Gateway references
	VpnGateway          string `gorm:"column:vpn_gateway;type:text" json:"vpnGateway"`
	TargetVpnGateway    string `gorm:"column:target_vpn_gateway;type:text" json:"targetVpnGateway"`
	VpnGatewayInterface int32  `gorm:"column:vpn_gateway_interface" json:"vpnGatewayInterface"`

	// JSONB fields
	LocalTrafficSelectorJSON  jsonb.JSON `gorm:"column:local_traffic_selector_json;type:jsonb" json:"localTrafficSelector"`
	RemoteTrafficSelectorJSON jsonb.JSON `gorm:"column:remote_traffic_selector_json;type:jsonb" json:"remoteTrafficSelector"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"projectId"`
	CollectedAt time.Time `gorm:"column:collected_at;not null" json:"collectedAt"`
}

func (GCPComputeVpnTunnel) TableName() string {
	return "bronze_history.gcp_compute_vpn_tunnels"
}

// GCPComputeVpnTunnelLabel stores historical versions of VPN tunnel labels.
// Links via VpnTunnelHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeVpnTunnelLabel struct {
	HistoryID          uint `gorm:"primaryKey"`
	VpnTunnelHistoryID uint `gorm:"column:vpn_tunnel_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	// Label fields
	Key   string `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value string `gorm:"column:value;type:text" json:"value"`
}

func (GCPComputeVpnTunnelLabel) TableName() string {
	return "bronze_history.gcp_compute_vpn_tunnel_labels"
}
