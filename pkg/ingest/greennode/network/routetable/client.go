package routetable

import (
	"context"
	"fmt"

	"danny.vn/greennode"
	"danny.vn/greennode/auth"
	"danny.vn/greennode/option"
	networkv2 "danny.vn/greennode/services/network/v2"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
)

// Client wraps the GreenNode SDK for route table operations.
type Client struct {
	sdk *greennode.Client
}

// NewClient creates a GreenNode client with rate limiting.
func NewClient(ctx context.Context, configService *config.Service, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter, region, projectID string) (*Client, error) {
	cfg := greennode.Config{
		Region:    region,
		ProjectID: projectID,
	}

	if iamAuth != nil {
		cfg.IAMAuth = iamAuth
	} else {
		cfg.ClientID = configService.GreenNodeClientID()
		cfg.ClientSecret = configService.GreenNodeClientSecret()
	}

	sdk, err := greennode.NewClient(ctx, cfg,
		option.WithTransport(ratelimit.NewRateLimitedTransport(limiter, nil)),
	)
	if err != nil {
		return nil, fmt.Errorf("create greennode client: %w", err)
	}

	return &Client{sdk: sdk}, nil
}

// ListRouteTables lists all route tables, handling pagination.
// Routes are embedded in each RouteTable response — no separate API call needed.
func (c *Client) ListRouteTables(ctx context.Context) ([]*networkv2.RouteTable, error) {
	var all []*networkv2.RouteTable
	page := 1
	size := 50

	for {
		result, err := c.sdk.Network.ListRouteTables(ctx, &networkv2.ListRouteTablesRequest{
			Page: page,
			Size: size,
		})
		if err != nil {
			return nil, fmt.Errorf("list route tables page %d: %w", page, err)
		}

		all = append(all, result.Items...)

		if page >= result.TotalPage {
			break
		}
		page++
	}

	return all, nil
}
