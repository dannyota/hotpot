package globaladdress

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for global addresses.
type Client struct {
	globalAddressesClient *compute.GlobalAddressesClient
}

// NewClient creates a new GCP Compute global address client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	globalAddressesClient, err := compute.NewGlobalAddressesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create global addresses client: %w", err)
	}

	return &Client{
		globalAddressesClient: globalAddressesClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.globalAddressesClient != nil {
		return c.globalAddressesClient.Close()
	}
	return nil
}

// ListGlobalAddresses lists all global addresses in a project.
func (c *Client) ListGlobalAddresses(ctx context.Context, projectID string) ([]*computepb.Address, error) {
	req := &computepb.ListGlobalAddressesRequest{
		Project: projectID,
	}

	var addresses []*computepb.Address
	it := c.globalAddressesClient.List(ctx, req)

	for {
		addr, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list global addresses in project %s: %w", projectID, err)
		}

		addresses = append(addresses, addr)
	}

	return addresses, nil
}
