package network

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"

	"hotpot/pkg/base/models/bronze"
)

// ConvertNetwork converts a GCP API Network to a Bronze model.
// Preserves raw API data with minimal transformation.
func ConvertNetwork(n *computepb.Network, projectID string, collectedAt time.Time) (bronze.GCPComputeNetwork, error) {
	network := bronze.GCPComputeNetwork{
		ResourceID:                            fmt.Sprintf("%d", n.GetId()),
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
		network.RoutingMode = n.RoutingConfig.GetRoutingMode()
	}

	// Convert subnetworks array to JSONB (nil → SQL NULL, data → JSON bytes)
	if n.Subnetworks != nil {
		var err error
		network.SubnetworksJSON, err = json.Marshal(n.Subnetworks)
		if err != nil {
			return bronze.GCPComputeNetwork{}, fmt.Errorf("failed to marshal subnetworks for network %s: %w", n.GetName(), err)
		}
	}

	// Convert peerings to separate table
	network.Peerings = ConvertPeerings(n.Peerings)

	return network, nil
}

// ConvertPeerings converts network peerings from GCP API to Bronze models.
func ConvertPeerings(peerings []*computepb.NetworkPeering) []bronze.GCPComputeNetworkPeering {
	if len(peerings) == 0 {
		return nil
	}

	result := make([]bronze.GCPComputeNetworkPeering, 0, len(peerings))
	for _, p := range peerings {
		result = append(result, bronze.GCPComputeNetworkPeering{
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
