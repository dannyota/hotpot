package subnetwork

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for subnetworks.
type Client struct {
	subnetworksClient *compute.SubnetworksClient
}

// NewClient creates a new GCP Compute subnetwork client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	subnetworksClient, err := compute.NewSubnetworksRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create subnetworks client: %w", err)
	}

	return &Client{
		subnetworksClient: subnetworksClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.subnetworksClient != nil {
		return c.subnetworksClient.Close()
	}
	return nil
}

// ListSubnetworks lists all subnetworks in a project using aggregated list.
// Returns subnetworks from all regions.
func (c *Client) ListSubnetworks(ctx context.Context, projectID string) ([]*computepb.Subnetwork, error) {
	req := &computepb.AggregatedListSubnetworksRequest{
		Project: projectID,
	}

	var subnetworks []*computepb.Subnetwork
	it := c.subnetworksClient.AggregatedList(ctx, req)

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list subnetworks in project %s: %w", projectID, err)
		}

		subnetworks = append(subnetworks, pair.Value.Subnetworks...)
	}

	return subnetworks, nil
}

// ListSubnetworksInRegion lists subnetworks in a specific region.
func (c *Client) ListSubnetworksInRegion(ctx context.Context, projectID, region string) ([]*computepb.Subnetwork, error) {
	req := &computepb.ListSubnetworksRequest{
		Project: projectID,
		Region:  region,
	}

	var subnetworks []*computepb.Subnetwork
	it := c.subnetworksClient.List(ctx, req)

	for {
		subnet, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list subnetworks in region %s: %w", region, err)
		}

		subnetworks = append(subnetworks, subnet)
	}

	return subnetworks, nil
}
