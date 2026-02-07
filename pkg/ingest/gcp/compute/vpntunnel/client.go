package vpntunnel

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for VPN tunnels.
type Client struct {
	vpnTunnelsClient *compute.VpnTunnelsClient
}

// NewClient creates a new GCP Compute VPN tunnel client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	vpnTunnelsClient, err := compute.NewVpnTunnelsRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create vpn tunnels client: %w", err)
	}

	return &Client{
		vpnTunnelsClient: vpnTunnelsClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.vpnTunnelsClient != nil {
		return c.vpnTunnelsClient.Close()
	}
	return nil
}

// ListVpnTunnels lists all VPN tunnels in a project using aggregated list.
// Returns VPN tunnels from all regions.
func (c *Client) ListVpnTunnels(ctx context.Context, projectID string) ([]*computepb.VpnTunnel, error) {
	req := &computepb.AggregatedListVpnTunnelsRequest{
		Project: projectID,
	}

	var vpnTunnels []*computepb.VpnTunnel
	it := c.vpnTunnelsClient.AggregatedList(ctx, req)

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list vpn tunnels in project %s: %w", projectID, err)
		}

		vpnTunnels = append(vpnTunnels, pair.Value.VpnTunnels...)
	}

	return vpnTunnels, nil
}
