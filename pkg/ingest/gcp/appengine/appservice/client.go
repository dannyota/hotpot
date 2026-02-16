package appservice

import (
	"context"
	"fmt"

	appengine "cloud.google.com/go/appengine/apiv1"
	"cloud.google.com/go/appengine/apiv1/appenginepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the GCP App Engine API for services.
type Client struct {
	servicesClient *appengine.ServicesClient
}

// NewClient creates a new App Engine services client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	servicesClient, err := appengine.NewServicesClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create App Engine services client: %w", err)
	}
	return &Client{servicesClient: servicesClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.servicesClient != nil {
		return c.servicesClient.Close()
	}
	return nil
}

// ListServices lists all App Engine services for a project.
func (c *Client) ListServices(ctx context.Context, projectID string) ([]*appenginepb.Service, error) {
	req := &appenginepb.ListServicesRequest{
		Parent: "apps/" + projectID,
	}

	var services []*appenginepb.Service
	it := c.servicesClient.ListServices(ctx, req)
	for {
		svc, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list services for project %s: %w", projectID, err)
		}
		services = append(services, svc)
	}
	return services, nil
}
