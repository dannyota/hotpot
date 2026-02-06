package targetinstance

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for target instances.
type Client struct {
	targetInstancesClient *compute.TargetInstancesClient
}

// NewClient creates a new GCP Compute target instances client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	tiClient, err := compute.NewTargetInstancesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create target instances client: %w", err)
	}

	return &Client{
		targetInstancesClient: tiClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.targetInstancesClient != nil {
		return c.targetInstancesClient.Close()
	}
	return nil
}

// ListTargetInstances lists all target instances in a project using aggregated list.
// Returns target instances from all zones.
func (c *Client) ListTargetInstances(ctx context.Context, projectID string) ([]*computepb.TargetInstance, error) {
	req := &computepb.AggregatedListTargetInstancesRequest{
		Project: projectID,
	}

	var targetInstances []*computepb.TargetInstance
	it := c.targetInstancesClient.AggregatedList(ctx, req)

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list target instances in project %s: %w", projectID, err)
		}

		targetInstances = append(targetInstances, pair.Value.TargetInstances...)
	}

	return targetInstances, nil
}
