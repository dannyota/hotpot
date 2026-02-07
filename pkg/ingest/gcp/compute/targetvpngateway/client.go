package targetvpngateway

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for Classic VPN gateways.
type Client struct {
	targetVpnGatewaysClient *compute.TargetVpnGatewaysClient
}

// NewClient creates a new GCP Compute Classic VPN gateway client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	targetVpnGatewaysClient, err := compute.NewTargetVpnGatewaysRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create target vpn gateways client: %w", err)
	}

	return &Client{
		targetVpnGatewaysClient: targetVpnGatewaysClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.targetVpnGatewaysClient != nil {
		return c.targetVpnGatewaysClient.Close()
	}
	return nil
}

// ListTargetVpnGateways lists all Classic VPN gateways in a project using aggregated list.
// Returns target VPN gateways from all regions.
func (c *Client) ListTargetVpnGateways(ctx context.Context, projectID string) ([]*computepb.TargetVpnGateway, error) {
	req := &computepb.AggregatedListTargetVpnGatewaysRequest{
		Project: projectID,
	}

	var targetVpnGateways []*computepb.TargetVpnGateway
	it := c.targetVpnGatewaysClient.AggregatedList(ctx, req)

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list target vpn gateways in project %s: %w", projectID, err)
		}

		targetVpnGateways = append(targetVpnGateways, pair.Value.TargetVpnGateways...)
	}

	return targetVpnGateways, nil
}
