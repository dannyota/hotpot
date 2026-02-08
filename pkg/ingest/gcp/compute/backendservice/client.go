package backendservice

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for backend services.
type Client struct {
	backendServicesClient *compute.BackendServicesClient
}

// NewClient creates a new GCP Compute backend services client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	c, err := compute.NewBackendServicesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create backend services client: %w", err)
	}
	return &Client{backendServicesClient: c}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.backendServicesClient != nil {
		return c.backendServicesClient.Close()
	}
	return nil
}

// ListBackendServices lists all backend services in a project using aggregated list.
// Returns backend services from all regions.
func (c *Client) ListBackendServices(ctx context.Context, projectID string) ([]*computepb.BackendService, error) {
	req := &computepb.AggregatedListBackendServicesRequest{
		Project: projectID,
	}

	var services []*computepb.BackendService
	it := c.backendServicesClient.AggregatedList(ctx, req)

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list backend services in project %s: %w", projectID, err)
		}
		services = append(services, pair.Value.BackendServices...)
	}

	return services, nil
}
