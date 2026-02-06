package bronze_history

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeNetwork stores historical snapshots of GCP VPC networks.
// Uses ResourceID for lookup (has API ID), with valid_from/valid_to for time range.
type GCPComputeNetwork struct {
	HistoryID uint `gorm:"primaryKey"`

	// Link by API ID (Network has unique API ID)
	ResourceID string `gorm:"column:resource_id;type:varchar(255);not null;index" json:"id"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"` // null = current

	// All network fields (same as bronze.GCPComputeNetwork)
	Name              string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description       string `gorm:"column:description;type:text" json:"description"`
	SelfLink          string `gorm:"column:self_link;type:text" json:"selfLink"`
	CreationTimestamp string `gorm:"column:creation_timestamp;type:varchar(50)" json:"creationTimestamp"`

	// Network configuration
	AutoCreateSubnetworks                 bool   `gorm:"column:auto_create_subnetworks" json:"autoCreateSubnetworks"`
	Mtu                                   int    `gorm:"column:mtu" json:"mtu"`
	RoutingMode                           string `gorm:"column:routing_mode;type:varchar(50)" json:"routingMode"`
	NetworkFirewallPolicyEnforcementOrder string `gorm:"column:firewall_policy_enforcement_order;type:varchar(50)" json:"networkFirewallPolicyEnforcementOrder"`

	// IPv6 configuration
	EnableUlaInternalIpv6 bool   `gorm:"column:enable_ula_internal_ipv6" json:"enableUlaInternalIpv6"`
	InternalIpv6Range     string `gorm:"column:internal_ipv6_range;type:varchar(50)" json:"internalIpv6Range"`

	// Gateway
	GatewayIpv4 string `gorm:"column:gateway_ipv4;type:varchar(50)" json:"gatewayIPv4"`

	// SubnetworksJSON
	SubnetworksJSON jsonb.JSON `gorm:"column:subnetworks_json;type:jsonb" json:"subnetworks"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"projectId"`
	CollectedAt time.Time `gorm:"column:collected_at;not null" json:"collectedAt"`
}

func (GCPComputeNetwork) TableName() string {
	return "bronze_history.gcp_compute_networks"
}

// GCPComputeNetworkPeering stores historical snapshots of network peerings.
// Links via NetworkHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeNetworkPeering struct {
	HistoryID        uint `gorm:"primaryKey"`
	NetworkHistoryID uint `gorm:"column:network_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	// All peering fields (same as bronze.GCPComputeNetworkPeering)
	Name                           string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Network                        string `gorm:"column:network;type:text" json:"network"`
	State                          string `gorm:"column:state;type:varchar(50)" json:"state"`
	StateDetails                   string `gorm:"column:state_details;type:text" json:"stateDetails"`
	ExportCustomRoutes             bool   `gorm:"column:export_custom_routes" json:"exportCustomRoutes"`
	ImportCustomRoutes             bool   `gorm:"column:import_custom_routes" json:"importCustomRoutes"`
	ExportSubnetRoutesWithPublicIp bool   `gorm:"column:export_subnet_routes_with_public_ip" json:"exportSubnetRoutesWithPublicIp"`
	ImportSubnetRoutesWithPublicIp bool   `gorm:"column:import_subnet_routes_with_public_ip" json:"importSubnetRoutesWithPublicIp"`
	ExchangeSubnetRoutes           bool   `gorm:"column:exchange_subnet_routes" json:"exchangeSubnetRoutes"`
	StackType                      string `gorm:"column:stack_type;type:varchar(50)" json:"stackType"`
	PeerMtu                        int    `gorm:"column:peer_mtu" json:"peerMtu"`
	AutoCreateRoutes               bool   `gorm:"column:auto_create_routes" json:"autoCreateRoutes"`
}

func (GCPComputeNetworkPeering) TableName() string {
	return "bronze_history.gcp_compute_network_peerings"
}
