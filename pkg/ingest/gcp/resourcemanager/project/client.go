package project

import (
	"context"
	"fmt"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Cloud Resource Manager API for projects.
type Client struct {
	projectsClient *resourcemanager.ProjectsClient
}

// NewClient creates a new GCP Resource Manager projects client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	projectsClient, err := resourcemanager.NewProjectsClient(ctx, opts...)
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

// SearchProjects searches for all projects accessible by the service account.
// Returns projects where the caller has resourcemanager.projects.get permission.
func (c *Client) SearchProjects(ctx context.Context) ([]*resourcemanagerpb.Project, error) {
	req := &resourcemanagerpb.SearchProjectsRequest{}

	var projects []*resourcemanagerpb.Project
	it := c.projectsClient.SearchProjects(ctx, req)

	for {
		project, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to search projects: %w", err)
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// GetProject retrieves a specific project by name (e.g., "projects/123456").
func (c *Client) GetProject(ctx context.Context, name string) (*resourcemanagerpb.Project, error) {
	req := &resourcemanagerpb.GetProjectRequest{
		Name: name,
	}

	project, err := c.projectsClient.GetProject(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", name, err)
	}

	return project, nil
}
