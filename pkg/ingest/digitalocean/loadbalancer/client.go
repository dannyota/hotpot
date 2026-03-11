package loadbalancer

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
)

// Client wraps the DigitalOcean Load Balancers API with pagination.
type Client struct {
	godoClient *godo.Client
}

// NewClient creates a new DigitalOcean Load Balancer client.
func NewClient(godoClient *godo.Client) *Client {
	return &Client{godoClient: godoClient}
}

// ListAll fetches all load balancers using page-based pagination.
func (c *Client) ListAll(ctx context.Context) ([]godo.LoadBalancer, error) {
	var all []godo.LoadBalancer
	opt := &godo.ListOptions{Page: 1, PerPage: 200}

	for {
		lbs, resp, err := c.godoClient.LoadBalancers.List(ctx, opt)
		if err != nil {
			return nil, fmt.Errorf("list load balancers (page %d): %w", opt.Page, err)
		}

		all = append(all, lbs...)

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		opt.Page++
	}

	return all, nil
}
