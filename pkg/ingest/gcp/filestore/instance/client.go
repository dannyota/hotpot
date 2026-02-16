package instance

import (
	"context"
	"fmt"

	filestore "cloud.google.com/go/filestore/apiv1"
	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the GCP Filestore API for instances.
type Client struct {
	fsClient *filestore.CloudFilestoreManagerClient
}

// NewClient creates a new GCP Filestore client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	fsClient, err := filestore.NewCloudFilestoreManagerClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Filestore client: %w", err)
	}

	return &Client{fsClient: fsClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.fsClient != nil {
		return c.fsClient.Close()
	}
	return nil
}

// ListInstances lists all Filestore instances in a project across all locations.
func (c *Client) ListInstances(ctx context.Context, projectID string) ([]*filestorepb.Instance, error) {
	var instances []*filestorepb.Instance

	parent := "projects/" + projectID + "/locations/-"
	req := &filestorepb.ListInstancesRequest{Parent: parent}

	it := c.fsClient.ListInstances(ctx, req)
	for {
		inst, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list Filestore instances in project %s: %w", projectID, err)
		}
		instances = append(instances, inst)
	}

	return instances, nil
}
