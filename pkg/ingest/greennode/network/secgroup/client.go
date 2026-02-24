package secgroup

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

// Client wraps the GreenNode SDK for security group operations.
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

// ListSecgroups lists all security groups.
func (c *Client) ListSecgroups(ctx context.Context) ([]*networkv2.Secgroup, error) {
	result, err := c.sdk.Network.ListSecgroup(ctx, &networkv2.ListSecgroupRequest{})
	if err != nil {
		return nil, fmt.Errorf("list secgroups: %w", err)
	}
	return result.Items, nil
}

// ListSecgroupRulesBySecgroupID lists all rules for a security group.
func (c *Client) ListSecgroupRulesBySecgroupID(ctx context.Context, secgroupID string) ([]*networkv2.SecgroupRule, error) {
	result, err := c.sdk.Network.ListSecgroupRulesBySecgroupID(ctx, &networkv2.ListSecgroupRulesBySecgroupIDRequest{
		SecgroupID: secgroupID,
	})
	if err != nil {
		return nil, fmt.Errorf("list secgroup rules for %s: %w", secgroupID, err)
	}
	return result.Items, nil
}
