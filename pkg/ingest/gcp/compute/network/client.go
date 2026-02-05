package network

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for networks.
type Client struct {
	networksClient *compute.NetworksClient
}

// NewClient creates a new GCP Compute network client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	networksClient, err := compute.NewNetworksRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create networks client: %w", err)
	}

	return &Client{
		networksClient: networksClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.networksClient != nil {
		return c.networksClient.Close()
	}
	return nil
}

// ListNetworks lists all networks in a project.
// Networks are global resources (not regional/zonal).
func (c *Client) ListNetworks(ctx context.Context, projectID string) ([]*computepb.Network, error) {
	req := &computepb.ListNetworksRequest{
		Project: projectID,
	}

	var networks []*computepb.Network
	it := c.networksClient.List(ctx, req)

	for {
		network, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list networks in project %s: %w", projectID, err)
		}

		networks = append(networks, network)
	}

	return networks, nil
}
