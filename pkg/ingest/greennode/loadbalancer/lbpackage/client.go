package lbpackage

import (
	"context"
	"fmt"

	"danny.vn/gnode"
	"danny.vn/gnode/auth"
	"danny.vn/gnode/option"
	lbv2 "danny.vn/gnode/services/loadbalancer/v2"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
)

// Client wraps the GreenNode SDK for load balancer package operations.
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

// ListPackages lists all load balancer packages.
func (c *Client) ListPackages(ctx context.Context) ([]*lbv2.LoadBalancerPackage, error) {
	result, err := c.sdk.LoadBalancer.ListLoadBalancerPackages(ctx, lbv2.NewListLoadBalancerPackagesRequest())
	if err != nil {
		return nil, fmt.Errorf("list load balancer packages: %w", err)
	}
	return result.Items, nil
}
