package targettcpproxy

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for target TCP proxies.
type Client struct {
	targetTcpProxiesClient *compute.TargetTcpProxiesClient
}

// NewClient creates a new GCP Compute target TCP proxies client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	ttpClient, err := compute.NewTargetTcpProxiesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create target TCP proxies client: %w", err)
	}

	return &Client{
		targetTcpProxiesClient: ttpClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.targetTcpProxiesClient != nil {
		return c.targetTcpProxiesClient.Close()
	}
	return nil
}

// ListTargetTcpProxies lists all target TCP proxies in a project (global resource).
func (c *Client) ListTargetTcpProxies(ctx context.Context, projectID string) ([]*computepb.TargetTcpProxy, error) {
	req := &computepb.ListTargetTcpProxiesRequest{
		Project: projectID,
	}

	var proxies []*computepb.TargetTcpProxy
	it := c.targetTcpProxiesClient.List(ctx, req)

	for {
		proxy, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list target TCP proxies in project %s: %w", projectID, err)
		}

		proxies = append(proxies, proxy)
	}

	return proxies, nil
}
