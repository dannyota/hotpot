package server

import (
	"context"
	"fmt"

	"danny.vn/greennode"
	"danny.vn/greennode/option"
	computev2 "danny.vn/greennode/services/compute/v2"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
)

// Client wraps the GreenNode SDK for server operations.
type Client struct {
	sdk *greennode.Client
}

// NewClient creates a GreenNode client with rate limiting.
func NewClient(ctx context.Context, configService *config.Service, limiter ratelimit.Limiter, region string) (*Client, error) {
	cfg := greennode.Config{
		Region:       region,
		ClientID:     configService.GreenNodeClientID(),
		ClientSecret: configService.GreenNodeClientSecret(),
		ProjectID:    configService.GreenNodeProjectID(),
	}

	sdk, err := greennode.NewClient(ctx, cfg,
		option.WithTransport(ratelimit.NewRateLimitedTransport(limiter, nil)),
	)
	if err != nil {
		return nil, fmt.Errorf("create greennode client: %w", err)
	}

	return &Client{sdk: sdk}, nil
}

// ListServers lists all servers, handling pagination.
func (c *Client) ListServers(ctx context.Context) ([]*computev2.Server, error) {
	var allServers []*computev2.Server
	page := 1
	size := 50

	for {
		result, err := c.sdk.Compute.ListServers(ctx, computev2.NewListServersRequest(page, size))
		if err != nil {
			return nil, fmt.Errorf("list servers page %d: %w", page, err)
		}

		allServers = append(allServers, result.Items...)

		if page >= result.TotalPage {
			break
		}
		page++
	}

	return allServers, nil
}
