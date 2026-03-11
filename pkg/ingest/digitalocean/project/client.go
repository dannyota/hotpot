package project

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
)

// Client wraps the DigitalOcean Projects API with pagination.
type Client struct {
	godoClient *godo.Client
}

// NewClient creates a new DigitalOcean Project client.
func NewClient(godoClient *godo.Client) *Client {
	return &Client{godoClient: godoClient}
}

// ListAllProjects fetches all projects using page-based pagination.
func (c *Client) ListAllProjects(ctx context.Context) ([]godo.Project, error) {
	var all []godo.Project
	opt := &godo.ListOptions{Page: 1, PerPage: 200}
	for {
		projects, resp, err := c.godoClient.Projects.List(ctx, opt)
		if err != nil {
			return nil, fmt.Errorf("list projects (page %d): %w", opt.Page, err)
		}
		all = append(all, projects...)
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}
		opt.Page++
	}
	return all, nil
}

// ListAllResources fetches all resources for a given project using page-based pagination.
func (c *Client) ListAllResources(ctx context.Context, projectID string) ([]godo.ProjectResource, error) {
	var all []godo.ProjectResource
	opt := &godo.ListOptions{Page: 1, PerPage: 200}
	for {
		resources, resp, err := c.godoClient.Projects.ListResources(ctx, projectID, opt)
		if err != nil {
			return nil, fmt.Errorf("list resources for project %s (page %d): %w", projectID, opt.Page, err)
		}
		all = append(all, resources...)
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}
		opt.Page++
	}
	return all, nil
}
