package glbresource

import (
	"context"
	"fmt"

	"danny.vn/gnode"
	"danny.vn/gnode/auth"
	"danny.vn/gnode/option"
	glbv1 "danny.vn/gnode/services/glb/v1"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
)

// Client wraps the GreenNode SDK for GLB operations.
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

// ListGlobalLoadBalancers lists all global load balancers, handling offset/limit pagination.
func (c *Client) ListGlobalLoadBalancers(ctx context.Context) ([]*glbv1.GlobalLoadBalancer, error) {
	var allGLBs []*glbv1.GlobalLoadBalancer
	offset := 0
	limit := 50

	for {
		result, err := c.sdk.GLB.ListGlobalLoadBalancers(ctx, glbv1.NewListGlobalLoadBalancersRequest(offset, limit))
		if err != nil {
			return nil, fmt.Errorf("list global load balancers offset %d: %w", offset, err)
		}

		allGLBs = append(allGLBs, result.Items...)

		if offset+limit >= result.Total {
			break
		}
		offset += limit
	}

	return allGLBs, nil
}

// ListGlobalListeners lists all listeners for a given load balancer.
func (c *Client) ListGlobalListeners(ctx context.Context, loadBalancerID string) ([]*glbv1.GlobalListener, error) {
	result, err := c.sdk.GLB.ListGlobalListeners(ctx, glbv1.NewListGlobalListenersRequest(loadBalancerID))
	if err != nil {
		return nil, fmt.Errorf("list global listeners for %s: %w", loadBalancerID, err)
	}
	return result.Items, nil
}

// ListGlobalPools lists all pools for a given load balancer.
func (c *Client) ListGlobalPools(ctx context.Context, loadBalancerID string) ([]*glbv1.GlobalPool, error) {
	result, err := c.sdk.GLB.ListGlobalPools(ctx, glbv1.NewListGlobalPoolsRequest(loadBalancerID))
	if err != nil {
		return nil, fmt.Errorf("list global pools for %s: %w", loadBalancerID, err)
	}
	return result.Items, nil
}

// ListGlobalPoolMembers lists all pool members for a given pool.
func (c *Client) ListGlobalPoolMembers(ctx context.Context, loadBalancerID, poolID string) ([]*glbv1.GlobalPoolMember, error) {
	result, err := c.sdk.GLB.ListGlobalPoolMembers(ctx, glbv1.NewListGlobalPoolMembersRequest(loadBalancerID, poolID))
	if err != nil {
		return nil, fmt.Errorf("list global pool members for pool %s: %w", poolID, err)
	}
	return result.Items, nil
}
