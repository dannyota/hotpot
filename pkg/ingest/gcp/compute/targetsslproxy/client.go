package targetsslproxy

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for target SSL proxies.
type Client struct {
	targetSslProxiesClient *compute.TargetSslProxiesClient
}

// NewClient creates a new GCP Compute target SSL proxies client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	tspClient, err := compute.NewTargetSslProxiesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create target SSL proxies client: %w", err)
	}

	return &Client{
		targetSslProxiesClient: tspClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.targetSslProxiesClient != nil {
		return c.targetSslProxiesClient.Close()
	}
	return nil
}

// ListTargetSslProxies lists all target SSL proxies in a project (global resource).
func (c *Client) ListTargetSslProxies(ctx context.Context, projectID string) ([]*computepb.TargetSslProxy, error) {
	req := &computepb.ListTargetSslProxiesRequest{
		Project: projectID,
	}

	var proxies []*computepb.TargetSslProxy
	it := c.targetSslProxiesClient.List(ctx, req)

	for {
		proxy, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list target SSL proxies in project %s: %w", projectID, err)
		}

		proxies = append(proxies, proxy)
	}

	return proxies, nil
}
