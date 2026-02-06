package address

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for regional addresses.
type Client struct {
	addressesClient *compute.AddressesClient
}

// NewClient creates a new GCP Compute address client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	addressesClient, err := compute.NewAddressesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create addresses client: %w", err)
	}

	return &Client{
		addressesClient: addressesClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.addressesClient != nil {
		return c.addressesClient.Close()
	}
	return nil
}

// ListAddresses lists all regional addresses in a project using aggregated list.
// Returns addresses from all regions.
func (c *Client) ListAddresses(ctx context.Context, projectID string) ([]*computepb.Address, error) {
	req := &computepb.AggregatedListAddressesRequest{
		Project: projectID,
	}

	var addresses []*computepb.Address
	it := c.addressesClient.AggregatedList(ctx, req)

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list addresses in project %s: %w", projectID, err)
		}

		addresses = append(addresses, pair.Value.Addresses...)
	}

	return addresses, nil
}
