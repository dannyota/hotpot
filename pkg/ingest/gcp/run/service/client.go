package service

import (
	"context"
	"fmt"

	run "cloud.google.com/go/run/apiv2"
	"cloud.google.com/go/run/apiv2/runpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the GCP Cloud Run API for services.
type Client struct {
	servicesClient *run.ServicesClient
}

// NewClient creates a new Cloud Run service client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	servicesClient, err := run.NewServicesClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Run services client: %w", err)
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

// ListServices lists all Cloud Run services in a project across all locations.
func (c *Client) ListServices(ctx context.Context, projectID string) ([]*runpb.Service, error) {
	req := &runpb.ListServicesRequest{
		Parent: "projects/" + projectID + "/locations/-",
	}

	var services []*runpb.Service
	it := c.servicesClient.ListServices(ctx, req)
	for {
		svc, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list Cloud Run services in project %s: %w", projectID, err)
		}
		services = append(services, svc)
	}
	return services, nil
}
