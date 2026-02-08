package targethttpproxy

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for target HTTP proxies.
type Client struct {
	targetHttpProxiesClient *compute.TargetHttpProxiesClient
}

// NewClient creates a new GCP Compute target HTTP proxies client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	thpClient, err := compute.NewTargetHttpProxiesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create target HTTP proxies client: %w", err)
	}

	return &Client{
		targetHttpProxiesClient: thpClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.targetHttpProxiesClient != nil {
		return c.targetHttpProxiesClient.Close()
	}
	return nil
}

// ListTargetHttpProxies lists all target HTTP proxies in a project using aggregated list.
func (c *Client) ListTargetHttpProxies(ctx context.Context, projectID string) ([]*computepb.TargetHttpProxy, error) {
	req := &computepb.AggregatedListTargetHttpProxiesRequest{
		Project: projectID,
	}

	var proxies []*computepb.TargetHttpProxy
	it := c.targetHttpProxiesClient.AggregatedList(ctx, req)

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list target HTTP proxies in project %s: %w", projectID, err)
		}

		proxies = append(proxies, pair.Value.TargetHttpProxies...)
	}

	return proxies, nil
}
