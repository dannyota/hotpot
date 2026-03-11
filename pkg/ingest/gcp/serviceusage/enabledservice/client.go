package enabledservice

import (
	"context"
	"fmt"

	serviceusage "cloud.google.com/go/serviceusage/apiv1"
	"cloud.google.com/go/serviceusage/apiv1/serviceusagepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the GCP Service Usage API for enabled services.
type Client struct {
	suClient *serviceusage.Client
}

// NewClient creates a new GCP Service Usage client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	suClient, err := serviceusage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Service Usage client: %w", err)
	}

	return &Client{suClient: suClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.suClient != nil {
		return c.suClient.Close()
	}
	return nil
}

// ListEnabledServices lists all enabled services in a project.
func (c *Client) ListEnabledServices(ctx context.Context, projectID string) ([]*serviceusagepb.Service, error) {
	var services []*serviceusagepb.Service

	parent := "projects/" + projectID
	req := &serviceusagepb.ListServicesRequest{
		Parent: parent,
		Filter: "state:ENABLED",
	}

	it := c.suClient.ListServices(ctx, req)
	for {
		s, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list enabled services in project %s: %w", projectID, err)
		}
		services = append(services, s)
	}

	return services, nil
}
