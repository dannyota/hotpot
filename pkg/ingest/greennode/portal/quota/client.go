package quota

import (
	"context"
	"fmt"

	"danny.vn/greennode"
	"danny.vn/greennode/option"
	portalv2 "danny.vn/greennode/services/portal/v2"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
)

// Client wraps the GreenNode SDK for quota operations.
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

// ListQuotas lists all quota usage.
func (c *Client) ListQuotas(ctx context.Context) ([]*portalv2.Quota, error) {
	result, err := c.sdk.Portal.ListAllQuotaUsed(ctx)
	if err != nil {
		return nil, fmt.Errorf("list quotas: %w", err)
	}
	return result.Items, nil
}
