package vpc

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
)

// Client wraps the DigitalOcean VPCs API with pagination.
type Client struct {
	godoClient *godo.Client
}

// NewClient creates a new DigitalOcean VPC client.
func NewClient(godoClient *godo.Client) *Client {
	return &Client{godoClient: godoClient}
}

// ListAll fetches all VPCs using page-based pagination.
func (c *Client) ListAll(ctx context.Context) ([]*godo.VPC, error) {
	var allVPCs []*godo.VPC
	opt := &godo.ListOptions{Page: 1, PerPage: 200}

	for {
		vpcs, resp, err := c.godoClient.VPCs.List(ctx, opt)
		if err != nil {
			return nil, fmt.Errorf("list VPCs (page %d): %w", opt.Page, err)
		}

		allVPCs = append(allVPCs, vpcs...)

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		opt.Page++
	}

	return allVPCs, nil
}
