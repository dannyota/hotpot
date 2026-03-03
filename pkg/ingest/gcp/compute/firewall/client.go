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

// ListFirewallsPage fetches a single page of firewalls from GCP.
func (c *Client) ListFirewallsPage(ctx context.Context, projectID string, pageSize int, pageToken string) ([]*computepb.Firewall, string, error) {
	it := c.firewallsClient.List(ctx, &computepb.ListFirewallsRequest{
		Project: projectID,
	})
	p := iterator.NewPager(it, pageSize, pageToken)

	var firewalls []*computepb.Firewall
	nextToken, err := p.NextPage(&firewalls)
	if err != nil {
		return nil, "", fmt.Errorf("list firewalls page: %w", err)
	}

	return firewalls, nextToken, nil
}
