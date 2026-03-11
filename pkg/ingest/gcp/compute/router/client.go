package router

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for routers.
type Client struct {
	routersClient *compute.RoutersClient
}

// NewClient creates a new GCP Compute router client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	routersClient, err := compute.NewRoutersRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create routers client: %w", err)
	}

	return &Client{
		routersClient: routersClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.routersClient != nil {
		return c.routersClient.Close()
	}
	return nil
}

// ListRouters lists all routers in a project using aggregated list.
// Returns routers from all regions.
func (c *Client) ListRouters(ctx context.Context, projectID string) ([]*computepb.Router, error) {
	req := &computepb.AggregatedListRoutersRequest{
		Project: projectID,
	}

	var routers []*computepb.Router
	it := c.routersClient.AggregatedList(ctx, req)

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list routers in project %s: %w", projectID, err)
		}

		routers = append(routers, pair.Value.Routers...)
	}

	return routers, nil
}
