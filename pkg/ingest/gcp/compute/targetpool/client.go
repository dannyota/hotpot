package targetpool

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for target pools.
type Client struct {
	targetPoolsClient *compute.TargetPoolsClient
}

// NewClient creates a new GCP Compute target pools client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	tpClient, err := compute.NewTargetPoolsRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create target pools client: %w", err)
	}

	return &Client{
		targetPoolsClient: tpClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.targetPoolsClient != nil {
		return c.targetPoolsClient.Close()
	}
	return nil
}

// ListTargetPools lists all target pools in a project using aggregated list.
func (c *Client) ListTargetPools(ctx context.Context, projectID string) ([]*computepb.TargetPool, error) {
	req := &computepb.AggregatedListTargetPoolsRequest{
		Project: projectID,
	}

	var pools []*computepb.TargetPool
	it := c.targetPoolsClient.AggregatedList(ctx, req)

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list target pools in project %s: %w", projectID, err)
		}

		pools = append(pools, pair.Value.TargetPools...)
	}

	return pools, nil
}
