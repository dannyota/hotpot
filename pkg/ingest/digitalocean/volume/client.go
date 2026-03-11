package volume

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
)

// Client wraps the DigitalOcean Volumes API with pagination.
type Client struct {
	godoClient *godo.Client
}

// NewClient creates a new DigitalOcean Volume client.
func NewClient(godoClient *godo.Client) *Client {
	return &Client{godoClient: godoClient}
}

// ListAll fetches all volumes using page-based pagination.
func (c *Client) ListAll(ctx context.Context) ([]godo.Volume, error) {
	var allVolumes []godo.Volume
	params := &godo.ListVolumeParams{ListOptions: &godo.ListOptions{Page: 1, PerPage: 200}}

	for {
		volumes, resp, err := c.godoClient.Storage.ListVolumes(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("list volumes (page %d): %w", params.ListOptions.Page, err)
		}

		allVolumes = append(allVolumes, volumes...)

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		params.ListOptions.Page++
	}

	return allVolumes, nil
}
