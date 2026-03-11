package firewall

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
)

// Client wraps the DigitalOcean Firewalls API with pagination.
type Client struct {
	godoClient *godo.Client
}

// NewClient creates a new DigitalOcean Firewall client.
func NewClient(godoClient *godo.Client) *Client {
	return &Client{godoClient: godoClient}
}

// ListAll fetches all firewalls using page-based pagination.
func (c *Client) ListAll(ctx context.Context) ([]godo.Firewall, error) {
	var all []godo.Firewall
	opt := &godo.ListOptions{Page: 1, PerPage: 200}

	for {
		firewalls, resp, err := c.godoClient.Firewalls.List(ctx, opt)
		if err != nil {
			return nil, fmt.Errorf("list firewalls (page %d): %w", opt.Page, err)
		}

		all = append(all, firewalls...)

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		opt.Page++
	}

	return all, nil
}
