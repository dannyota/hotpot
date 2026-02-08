package network

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// NetworkData holds converted network data ready for Ent insertion.
type NetworkData struct {
	ID                                    string
	Name                                  string
	Description                           string
	SelfLink                              string
	CreationTimestamp                     string
	AutoCreateSubnetworks                 bool
	Mtu                                   int
	RoutingMode                           string
	NetworkFirewallPolicyEnforcementOrder string
	EnableUlaInternalIpv6                 bool
	InternalIpv6Range                     string
	GatewayIpv4                           string
	SubnetworksJSON                       json.RawMessage
	Peerings                              []PeeringData
	ProjectID                             string
	CollectedAt                           time.Time
}

// PeeringData holds converted peering data.
type PeeringData struct {
	Name                           string
	Network                        string
	State                          string
	StateDetails                   string
	ExportCustomRoutes             bool
	ImportCustomRoutes             bool
	ExportSubnetRoutesWithPublicIp bool
	ImportSubnetRoutesWithPublicIp bool
	ExchangeSubnetRoutes           bool
	StackType                      string
	PeerMtu                        int
	AutoCreateRoutes               bool
}

// ConvertNetwork converts a GCP API Network to Ent-compatible data.
// Preserves raw API data with minimal transformation.
func ConvertNetwork(n *computepb.Network, projectID string, collectedAt time.Time) (*NetworkData, error) {
	data := &NetworkData{
		ID:                                    fmt.Sprintf("%d", n.GetId()),
		Name:                                  n.GetName(),
		Description:                           n.GetDescription(),
		SelfLink:                              n.GetSelfLink(),
		CreationTimestamp:                     n.GetCreationTimestamp(),
		AutoCreateSubnetworks:                 n.GetAutoCreateSubnetworks(),
		Mtu:                                   int(n.GetMtu()),
		NetworkFirewallPolicyEnforcementOrder: n.GetNetworkFirewallPolicyEnforcementOrder(),
		EnableUlaInternalIpv6:                 n.GetEnableUlaInternalIpv6(),
		InternalIpv6Range:                     n.GetInternalIpv6Range(),
		GatewayIpv4:                           n.GetGatewayIPv4(),
		ProjectID:                             projectID,
		CollectedAt:                           collectedAt,
	}

	// Extract routing mode from routingConfig
	if n.RoutingConfig != nil {
		data.RoutingMode = n.RoutingConfig.GetRoutingMode()
	}

	// Convert subnetworks array to JSONB (nil → SQL NULL, data → JSON bytes)
	if n.Subnetworks != nil {
		var err error
		data.SubnetworksJSON, err = json.Marshal(n.Subnetworks)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal subnetworks for network %s: %w", n.GetName(), err)
		}
	}

	// Convert peerings to separate table
	data.Peerings = ConvertPeerings(n.Peerings)

	return data, nil
}

// ConvertPeerings converts network peerings from GCP API to data structs.
func ConvertPeerings(peerings []*computepb.NetworkPeering) []PeeringData {
	if len(peerings) == 0 {
		return nil
	}

	result := make([]PeeringData, 0, len(peerings))
	for _, p := range peerings {
		result = append(result, PeeringData{
			Name:                           p.GetName(),
			Network:                        p.GetNetwork(),
			State:                          p.GetState(),
			StateDetails:                   p.GetStateDetails(),
			ExportCustomRoutes:             p.GetExportCustomRoutes(),
			ImportCustomRoutes:             p.GetImportCustomRoutes(),
			ExportSubnetRoutesWithPublicIp: p.GetExportSubnetRoutesWithPublicIp(),
			ImportSubnetRoutesWithPublicIp: p.GetImportSubnetRoutesWithPublicIp(),
			ExchangeSubnetRoutes:           p.GetExchangeSubnetRoutes(),
			StackType:                      p.GetStackType(),
			PeerMtu:                        int(p.GetPeerMtu()),
			AutoCreateRoutes:               p.GetAutoCreateRoutes(),
		})
	}

	return result
}
