package application

import (
	"context"
	"fmt"

	appengine "cloud.google.com/go/appengine/apiv1"
	"cloud.google.com/go/appengine/apiv1/appenginepb"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Client wraps the GCP App Engine API for applications.
type Client struct {
	appsClient *appengine.ApplicationsClient
}

// NewClient creates a new App Engine applications client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	appsClient, err := appengine.NewApplicationsClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create App Engine applications client: %w", err)
	}
	return &Client{appsClient: appsClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.appsClient != nil {
		return c.appsClient.Close()
	}
	return nil
}

// GetApplication fetches the App Engine application for a project.
// Returns nil, nil if the application does not exist.
func (c *Client) GetApplication(ctx context.Context, projectID string) (*appenginepb.Application, error) {
	req := &appenginepb.GetApplicationRequest{
		Name: "apps/" + projectID,
	}

	app, err := c.appsClient.GetApplication(ctx, req)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get application for project %s: %w", projectID, err)
	}

	return app, nil
}
