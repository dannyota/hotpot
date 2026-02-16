package instance

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigtable"
	"google.golang.org/api/option"
)

// Client wraps the GCP Bigtable Instance Admin API for instances.
type Client struct {
	adminClient *bigtable.InstanceAdminClient
	projectID   string
}

// NewClient creates a new Bigtable instance admin client.
func NewClient(ctx context.Context, projectID string, opts ...option.ClientOption) (*Client, error) {
	adminClient, err := bigtable.NewInstanceAdminClient(ctx, projectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create bigtable instance admin client: %w", err)
	}
	return &Client{adminClient: adminClient, projectID: projectID}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.adminClient != nil {
		return c.adminClient.Close()
	}
	return nil
}

// ListInstances lists all Bigtable instances in the project.
func (c *Client) ListInstances(ctx context.Context) ([]*bigtable.InstanceInfo, error) {
	instances, err := c.adminClient.Instances(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list bigtable instances in project %s: %w", c.projectID, err)
	}
	return instances, nil
}
