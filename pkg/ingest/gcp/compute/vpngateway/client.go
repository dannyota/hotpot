package vpngateway

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for VPN gateways.
type Client struct {
	vpnGatewaysClient *compute.VpnGatewaysClient
}

// NewClient creates a new GCP Compute VPN gateway client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	vpnGatewaysClient, err := compute.NewVpnGatewaysRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create vpn gateways client: %w", err)
	}

	return &Client{
		vpnGatewaysClient: vpnGatewaysClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.vpnGatewaysClient != nil {
		return c.vpnGatewaysClient.Close()
	}
	return nil
}

// ListVpnGateways lists all VPN gateways in a project using aggregated list.
// Returns VPN gateways from all regions.
func (c *Client) ListVpnGateways(ctx context.Context, projectID string) ([]*computepb.VpnGateway, error) {
	req := &computepb.AggregatedListVpnGatewaysRequest{
		Project: projectID,
	}

	var vpnGateways []*computepb.VpnGateway
	it := c.vpnGatewaysClient.AggregatedList(ctx, req)

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list vpn gateways in project %s: %w", projectID, err)
		}

		vpnGateways = append(vpnGateways, pair.Value.VpnGateways...)
	}

	return vpnGateways, nil
}
