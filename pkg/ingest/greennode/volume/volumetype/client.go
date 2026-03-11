package volumetype

import (
	"context"
	"fmt"

	"danny.vn/gnode"
	"danny.vn/gnode/auth"
	"danny.vn/gnode/option"
	volumev1 "danny.vn/gnode/services/volume/v1"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
)

// Client wraps the GreenNode SDK for volume type operations.
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

// ListVolumeTypes lists all volume types across all zones.
func (c *Client) ListVolumeTypes(ctx context.Context) ([]*volumev1.VolumeType, error) {
	zonesResult, err := c.sdk.VolumeV1.GetVolumeTypeZones(ctx, &volumev1.GetVolumeTypeZonesRequest{})
	if err != nil {
		return nil, fmt.Errorf("list volume type zones: %w", err)
	}

	var all []*volumev1.VolumeType
	for _, zone := range zonesResult.VolumeTypeZones {
		result, err := c.sdk.VolumeV1.GetListVolumeTypes(ctx, &volumev1.GetListVolumeTypeRequest{
			VolumeTypeZoneID: zone.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("list volume types for zone %s: %w", zone.ID, err)
		}
		all = append(all, result.VolumeTypes...)
	}

	return all, nil
}
