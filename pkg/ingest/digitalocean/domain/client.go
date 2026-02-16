package domain

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
)

// Client wraps the DigitalOcean Domains API with pagination.
type Client struct {
	godoClient *godo.Client
}

// NewClient creates a new DigitalOcean Domain client.
func NewClient(godoClient *godo.Client) *Client {
	return &Client{godoClient: godoClient}
}

// ListAllDomains fetches all domains using page-based pagination.
func (c *Client) ListAllDomains(ctx context.Context) ([]godo.Domain, error) {
	var all []godo.Domain
	opt := &godo.ListOptions{Page: 1, PerPage: 200}
	for {
		domains, resp, err := c.godoClient.Domains.List(ctx, opt)
		if err != nil {
			return nil, fmt.Errorf("list domains (page %d): %w", opt.Page, err)
		}
		all = append(all, domains...)
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}
		opt.Page++
	}
	return all, nil
}

// ListAllRecords fetches all records for a given domain using page-based pagination.
func (c *Client) ListAllRecords(ctx context.Context, domainName string) ([]godo.DomainRecord, error) {
	var all []godo.DomainRecord
	opt := &godo.ListOptions{Page: 1, PerPage: 200}
	for {
		records, resp, err := c.godoClient.Domains.Records(ctx, domainName, opt)
		if err != nil {
			return nil, fmt.Errorf("list records for domain %s (page %d): %w", domainName, opt.Page, err)
		}
		all = append(all, records...)
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}
		opt.Page++
	}
	return all, nil
}
