package key

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
)

// Client wraps the DigitalOcean Keys API with pagination.
type Client struct {
	godoClient *godo.Client
}

// NewClient creates a new DigitalOcean SSH Key client.
func NewClient(godoClient *godo.Client) *Client {
	return &Client{godoClient: godoClient}
}

// ListAll fetches all SSH keys using page-based pagination.
func (c *Client) ListAll(ctx context.Context) ([]godo.Key, error) {
	var allKeys []godo.Key
	opt := &godo.ListOptions{Page: 1, PerPage: 200}

	for {
		keys, resp, err := c.godoClient.Keys.List(ctx, opt)
		if err != nil {
			return nil, fmt.Errorf("list keys (page %d): %w", opt.Page, err)
		}

		allKeys = append(allKeys, keys...)

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		opt.Page++
	}

	return allKeys, nil
}
