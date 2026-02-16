package instance

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

// Client wraps GCP Cloud SQL Admin API for instances.
type Client struct {
	service *sqladmin.Service
}

// NewClient creates a new GCP Cloud SQL instance client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	service, err := sqladmin.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create sqladmin service: %w", err)
	}

	return &Client{
		service: service,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	// sqladmin.Service doesn't have a Close method
	return nil
}

// ListInstances lists all Cloud SQL instances in a project.
func (c *Client) ListInstances(ctx context.Context, projectID string) ([]*sqladmin.DatabaseInstance, error) {
	var instances []*sqladmin.DatabaseInstance

	err := c.service.Instances.List(projectID).Pages(ctx, func(resp *sqladmin.InstancesListResponse) error {
		instances = append(instances, resp.Items...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list SQL instances in project %s: %w", projectID, err)
	}

	return instances, nil
}
