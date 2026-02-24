package lb

import (
	"context"
	"fmt"

	"danny.vn/greennode"
	"danny.vn/greennode/auth"
	"danny.vn/greennode/option"
	lbv2 "danny.vn/greennode/services/loadbalancer/v2"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
)

// Client wraps the GreenNode SDK for load balancer operations.
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

// ListLoadBalancers lists all load balancers, handling pagination.
func (c *Client) ListLoadBalancers(ctx context.Context) ([]*lbv2.LoadBalancer, error) {
	var allLBs []*lbv2.LoadBalancer
	page := 1
	size := 50

	for {
		result, err := c.sdk.LoadBalancer.ListLoadBalancers(ctx, lbv2.NewListLoadBalancersRequest(page, size))
		if err != nil {
			return nil, fmt.Errorf("list load balancers page %d: %w", page, err)
		}

		allLBs = append(allLBs, result.Items...)

		if page >= result.TotalPage {
			break
		}
		page++
	}

	return allLBs, nil
}

// ListListenersByLBID lists all listeners for a load balancer.
func (c *Client) ListListenersByLBID(ctx context.Context, lbID string) ([]*lbv2.Listener, error) {
	result, err := c.sdk.LoadBalancer.ListListenersByLoadBalancerID(ctx, lbv2.NewListListenersByLoadBalancerIDRequest(lbID))
	if err != nil {
		return nil, fmt.Errorf("list listeners for LB %s: %w", lbID, err)
	}
	return result.Items, nil
}

// ListPoolsByLBID lists all pools for a load balancer.
func (c *Client) ListPoolsByLBID(ctx context.Context, lbID string) ([]*lbv2.Pool, error) {
	result, err := c.sdk.LoadBalancer.ListPoolsByLoadBalancerID(ctx, lbv2.NewListPoolsByLoadBalancerIDRequest(lbID))
	if err != nil {
		return nil, fmt.Errorf("list pools for LB %s: %w", lbID, err)
	}
	return result.Items, nil
}
