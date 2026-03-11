package account

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
)

// Client wraps the DigitalOcean Account API.
type Client struct {
	godoClient *godo.Client
}

// NewClient creates a new DigitalOcean Account client.
func NewClient(godoClient *godo.Client) *Client {
	return &Client{godoClient: godoClient}
}

// Get fetches the DigitalOcean account.
func (c *Client) Get(ctx context.Context) (*godo.Account, error) {
	account, _, err := c.godoClient.Account.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}

	return account, nil
}
