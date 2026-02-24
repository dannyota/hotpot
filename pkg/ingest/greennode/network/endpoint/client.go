package endpoint

import (
	"context"
	"fmt"

	"danny.vn/greennode"
	"danny.vn/greennode/auth"
	"danny.vn/greennode/option"
	networkv1 "danny.vn/greennode/services/network/v1"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
)

// Client wraps the GreenNode SDK for endpoint operations.
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

// ListEndpoints lists all endpoints, handling pagination.
func (c *Client) ListEndpoints(ctx context.Context) ([]*networkv1.Endpoint, error) {
	var allEndpoints []*networkv1.Endpoint
	page := 1
	size := 50

	for {
		result, err := c.sdk.NetworkV1.ListEndpoints(ctx, &networkv1.ListEndpointsRequest{
			Page: page,
			Size: size,
		})
		if err != nil {
			return nil, fmt.Errorf("list endpoints page %d: %w", page, err)
		}

		allEndpoints = append(allEndpoints, result.Items...)

		if page >= result.TotalPage {
			break
		}
		page++
	}

	return allEndpoints, nil
}
