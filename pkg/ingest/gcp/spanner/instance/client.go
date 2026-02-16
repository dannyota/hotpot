package instance

import (
	"context"
	"fmt"

	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	"cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the GCP Spanner Instance Admin API.
type Client struct {
	instanceAdmin *instance.InstanceAdminClient
}

// NewClient creates a new Spanner instance admin client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	instanceAdmin, err := instance.NewInstanceAdminClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create spanner instance admin client: %w", err)
	}
	return &Client{instanceAdmin: instanceAdmin}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.instanceAdmin != nil {
		return c.instanceAdmin.Close()
	}
	return nil
}

// ListInstances lists all Spanner instances in a project.
func (c *Client) ListInstances(ctx context.Context, projectID string) ([]*instancepb.Instance, error) {
	req := &instancepb.ListInstancesRequest{
		Parent: "projects/" + projectID,
	}

	var instances []*instancepb.Instance
	it := c.instanceAdmin.ListInstances(ctx, req)
	for {
		inst, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list spanner instances in project %s: %w", projectID, err)
		}
		instances = append(instances, inst)
	}
	return instances, nil
}
