package volumetypezone

import (
	"context"
	"fmt"

	"danny.vn/greennode"
	"danny.vn/greennode/auth"
	"danny.vn/greennode/option"
	volumev1 "danny.vn/greennode/services/volume/v1"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
)

// Client wraps the GreenNode SDK for volume type zone operations.
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

// ListVolumeTypeZones lists all volume type zones.
func (c *Client) ListVolumeTypeZones(ctx context.Context) ([]*volumev1.VolumeTypeZone, error) {
	result, err := c.sdk.VolumeV1.GetVolumeTypeZones(ctx, &volumev1.GetVolumeTypeZonesRequest{})
	if err != nil {
		return nil, fmt.Errorf("list volume type zones: %w", err)
	}
	return result.VolumeTypeZones, nil
}
