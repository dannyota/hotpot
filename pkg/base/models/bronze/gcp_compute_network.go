package bronze

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeNetwork represents a GCP VPC network in the bronze layer.
// Fields preserve raw API response data from compute.networks.list.
type GCPComputeNetwork struct {
	// GCP API fields (json tag = original API field name for traceability)
	// ResourceID is the GCP API ID, used as primary key for linking.
	ResourceID        string `gorm:"primaryKey;column:resource_id;type:varchar(255)" json:"id"`
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

	// SubnetworksJSON contains list of subnetwork URLs in this network.
	//
	//	["projects/.../regions/.../subnetworks/subnet1", ...]
	SubnetworksJSON jsonb.JSON `gorm:"column:subnetworks_json;type:jsonb" json:"subnetworks"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"-"`
	CollectedAt time.Time `gorm:"column:collected_at;not null;index" json:"-"`

	// Relationships
	Peerings []GCPComputeNetworkPeering `gorm:"foreignKey:NetworkResourceID;references:ResourceID" json:"peerings,omitempty"`
}

func (GCPComputeNetwork) TableName() string {
	return "bronze.gcp_compute_networks"
}

// GCPComputeNetworkPeering represents a VPC network peering connection.
// Data from network.peerings[].
type GCPComputeNetworkPeering struct {
	ID                uint   `gorm:"primaryKey"`
	NetworkResourceID string `gorm:"column:network_resource_id;type:varchar(255);not null;index" json:"-"`

	// GCP API fields (json tag = original API field name)
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
	return "bronze.gcp_compute_network_peerings"
}
