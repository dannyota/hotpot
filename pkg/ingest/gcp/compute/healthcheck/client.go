package healthcheck

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for health checks.
type Client struct {
	healthChecksClient *compute.HealthChecksClient
}

// NewClient creates a new GCP Compute health check client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	healthChecksClient, err := compute.NewHealthChecksRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create health checks client: %w", err)
	}

	return &Client{
		healthChecksClient: healthChecksClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.healthChecksClient != nil {
		return c.healthChecksClient.Close()
	}
	return nil
}

// ListHealthChecks lists all health checks in a project using aggregated list.
// Returns health checks from all regions plus global.
func (c *Client) ListHealthChecks(ctx context.Context, projectID string) ([]*computepb.HealthCheck, error) {
	req := &computepb.AggregatedListHealthChecksRequest{
		Project: projectID,
	}

	var checks []*computepb.HealthCheck
	it := c.healthChecksClient.AggregatedList(ctx, req)

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list health checks in project %s: %w", projectID, err)
		}

		checks = append(checks, pair.Value.HealthChecks...)
	}

	return checks, nil
}
