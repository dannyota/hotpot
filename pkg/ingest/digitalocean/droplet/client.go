package droplet

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
)

// Client wraps the DigitalOcean Droplets API with pagination.
type Client struct {
	godoClient *godo.Client
}

// NewClient creates a new DigitalOcean Droplet client.
func NewClient(godoClient *godo.Client) *Client {
	return &Client{godoClient: godoClient}
}

// ListAll fetches all droplets using page-based pagination.
func (c *Client) ListAll(ctx context.Context) ([]godo.Droplet, error) {
	var all []godo.Droplet
	opt := &godo.ListOptions{Page: 1, PerPage: 200}

	for {
		droplets, resp, err := c.godoClient.Droplets.List(ctx, opt)
		if err != nil {
			return nil, fmt.Errorf("list droplets (page %d): %w", opt.Page, err)
		}

		all = append(all, droplets...)

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		opt.Page++
	}

	return all, nil
}
