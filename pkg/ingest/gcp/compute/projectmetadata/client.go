package projectmetadata

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for project metadata.
type Client struct {
	projectsClient *compute.ProjectsClient
}

// NewClient creates a new GCP Compute projects client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	projectsClient, err := compute.NewProjectsRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create projects client: %w", err)
	}

	return &Client{
		projectsClient: projectsClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.projectsClient != nil {
		return c.projectsClient.Close()
	}
	return nil
}

// GetProjectMetadata fetches project metadata for a single project.
// Returns a single *computepb.Project (one per project).
func (c *Client) GetProjectMetadata(ctx context.Context, projectID string) (*computepb.Project, error) {
	req := &computepb.GetProjectRequest{
		Project: projectID,
	}

	project, err := c.projectsClient.Get(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get project metadata for project %s: %w", projectID, err)
	}

	return project, nil
}
