package region

import (
	"context"
	"fmt"

	"danny.vn/greennode"
	"danny.vn/greennode/auth"
	"danny.vn/greennode/option"
	portalv2 "danny.vn/greennode/services/portal/v2"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
)

// Client wraps the GreenNode SDK for region operations.
type Client struct {
	sdk *greennode.Client
}

// NewClient creates a GreenNode client with rate limiting.
func NewClient(ctx context.Context, configService *config.Service, limiter ratelimit.Limiter, region, projectID string) (*Client, error) {
	cfg := greennode.Config{
		Region:    region,
		ProjectID: projectID,
	}

	if username := configService.GreenNodeUsername(); username != "" {
		iamAuth := &auth.IAMUserAuth{
			RootEmail: configService.GreenNodeRootEmail(),
			Username:  username,
			Password:  configService.GreenNodePassword(),
		}
		if totpSecret := configService.GreenNodeTOTPSecret(); totpSecret != "" {
			iamAuth.TOTP = &auth.SecretTOTP{Secret: totpSecret}
		}
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

// ListRegions lists all regions.
func (c *Client) ListRegions(ctx context.Context) ([]*portalv2.Region, error) {
	result, err := c.sdk.Portal.ListRegions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list regions: %w", err)
	}
	return result.Items, nil
}
