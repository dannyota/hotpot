package interconnect

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for interconnects.
type Client struct {
	interconnectsClient *compute.InterconnectsClient
}

// NewClient creates a new GCP Compute interconnect client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	interconnectsClient, err := compute.NewInterconnectsRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create interconnects client: %w", err)
	}

	return &Client{
		interconnectsClient: interconnectsClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.interconnectsClient != nil {
		return c.interconnectsClient.Close()
	}
	return nil
}

// ListInterconnects lists all interconnects in a project.
// Interconnects are global resources (not regional/zonal).
func (c *Client) ListInterconnects(ctx context.Context, projectID string) ([]*computepb.Interconnect, error) {
	req := &computepb.ListInterconnectsRequest{
		Project: projectID,
	}

	var interconnects []*computepb.Interconnect
	it := c.interconnectsClient.List(ctx, req)

	for {
		interconnect, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list interconnects in project %s: %w", projectID, err)
		}

		interconnects = append(interconnects, interconnect)
	}

	return interconnects, nil
}
