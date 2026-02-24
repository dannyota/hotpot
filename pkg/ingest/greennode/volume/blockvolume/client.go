package blockvolume

import (
	"context"
	"fmt"

	"danny.vn/greennode"
	"danny.vn/greennode/auth"
	"danny.vn/greennode/option"
	volumev2 "danny.vn/greennode/services/volume/v2"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
)

// Client wraps the GreenNode SDK for block volume operations.
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

// ListBlockVolumes lists all block volumes.
func (c *Client) ListBlockVolumes(ctx context.Context) ([]*volumev2.Volume, error) {
	result, err := c.sdk.Volume.ListBlockVolumes(ctx, &volumev2.ListBlockVolumesRequest{
		Page: 1,
		Size: 10000,
	})
	if err != nil {
		return nil, fmt.Errorf("list block volumes: %w", err)
	}
	return result.Items, nil
}

// ListSnapshotsByBlockVolumeID lists all snapshots for a block volume.
func (c *Client) ListSnapshotsByBlockVolumeID(ctx context.Context, blockVolumeID string) ([]*volumev2.Snapshot, error) {
	result, err := c.sdk.Volume.ListSnapshotsByBlockVolumeID(ctx, &volumev2.ListSnapshotsByBlockVolumeIDRequest{
		BlockVolumeID: blockVolumeID,
		Page:          1,
		Size:          10000,
	})
	if err != nil {
		return nil, fmt.Errorf("list snapshots for volume %s: %w", blockVolumeID, err)
	}
	return result.Items, nil
}
