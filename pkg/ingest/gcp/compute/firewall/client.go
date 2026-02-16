package firewall

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for firewalls.
type Client struct {
	firewallsClient *compute.FirewallsClient
}

// NewClient creates a new GCP Compute firewall client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	firewallsClient, err := compute.NewFirewallsRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create firewalls client: %w", err)
	}

	return &Client{
		firewallsClient: firewallsClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.firewallsClient != nil {
		return c.firewallsClient.Close()
	}
	return nil
}

// ListFirewalls lists all firewalls in a project.
// Firewalls are global resources (not regional/zonal).
func (c *Client) ListFirewalls(ctx context.Context, projectID string) ([]*computepb.Firewall, error) {
	req := &computepb.ListFirewallsRequest{
		Project: projectID,
	}

	var firewalls []*computepb.Firewall
	it := c.firewallsClient.List(ctx, req)

	for {
		firewall, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list firewalls in project %s: %w", projectID, err)
		}

		firewalls = append(firewalls, firewall)
	}

	return firewalls, nil
}
