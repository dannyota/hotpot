package targethttpsproxy

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for target HTTPS proxies.
type Client struct {
	targetHttpsProxiesClient *compute.TargetHttpsProxiesClient
}

// NewClient creates a new GCP Compute target HTTPS proxies client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	thspClient, err := compute.NewTargetHttpsProxiesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create target HTTPS proxies client: %w", err)
	}

	return &Client{
		targetHttpsProxiesClient: thspClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.targetHttpsProxiesClient != nil {
		return c.targetHttpsProxiesClient.Close()
	}
	return nil
}

// ListTargetHttpsProxies lists all target HTTPS proxies in a project using aggregated list.
func (c *Client) ListTargetHttpsProxies(ctx context.Context, projectID string) ([]*computepb.TargetHttpsProxy, error) {
	req := &computepb.AggregatedListTargetHttpsProxiesRequest{
		Project: projectID,
	}

	var proxies []*computepb.TargetHttpsProxy
	it := c.targetHttpsProxiesClient.AggregatedList(ctx, req)

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list target HTTPS proxies in project %s: %w", projectID, err)
		}

		proxies = append(proxies, pair.Value.TargetHttpsProxies...)
	}

	return proxies, nil
}
