package instance

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for instances.
type Client struct {
	instancesClient *compute.InstancesClient
}

// NewClient creates a new GCP Compute instance client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	instancesClient, err := compute.NewInstancesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create instances client: %w", err)
	}

	return &Client{
		instancesClient: instancesClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.instancesClient != nil {
		return c.instancesClient.Close()
	}
	return nil
}

// ListInstances lists all instances in a project using aggregated list.
// Returns instances from all zones.
func (c *Client) ListInstances(ctx context.Context, projectID string) ([]*computepb.Instance, error) {
	req := &computepb.AggregatedListInstancesRequest{
		Project: projectID,
	}

	var instances []*computepb.Instance
	it := c.instancesClient.AggregatedList(ctx, req)

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list instances in project %s: %w", projectID, err)
		}

		instances = append(instances, pair.Value.Instances...)
	}

	return instances, nil
}

// ListInstancesInZone lists instances in a specific zone.
func (c *Client) ListInstancesInZone(ctx context.Context, projectID, zone string) ([]*computepb.Instance, error) {
	req := &computepb.ListInstancesRequest{
		Project: projectID,
		Zone:    zone,
	}

	var instances []*computepb.Instance
	it := c.instancesClient.List(ctx, req)

	for {
		instance, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list instances in zone %s: %w", zone, err)
		}

		instances = append(instances, instance)
	}

	return instances, nil
}
