package blockvolume

import (
	"context"
	"fmt"

	"danny.vn/gnode"
	"danny.vn/gnode/auth"
	"danny.vn/gnode/option"
	volumev2 "danny.vn/gnode/services/volume/v2"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
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

// ListBlockVolumes lists all block volumes, handling pagination.
func (c *Client) ListBlockVolumes(ctx context.Context) ([]*volumev2.Volume, error) {
	var allVolumes []*volumev2.Volume
	page := 1
	size := 50

	for {
		result, err := c.sdk.Volume.ListBlockVolumes(ctx, &volumev2.ListBlockVolumesRequest{
			Page: page,
			Size: size,
		})
		if err != nil {
			return nil, fmt.Errorf("list block volumes page %d: %w", page, err)
		}

		allVolumes = append(allVolumes, result.Items...)

		if page >= result.TotalPage {
			break
		}
		page++
	}

	return allVolumes, nil
}

// ListSnapshotsByBlockVolumeID lists all snapshots for a block volume, handling pagination.
func (c *Client) ListSnapshotsByBlockVolumeID(ctx context.Context, blockVolumeID string) ([]*volumev2.Snapshot, error) {
	var allSnapshots []*volumev2.Snapshot
	page := 1
	size := 50

	for {
		result, err := c.sdk.Volume.ListSnapshotsByBlockVolumeID(ctx, &volumev2.ListSnapshotsByBlockVolumeIDRequest{
			BlockVolumeID: blockVolumeID,
			Page:          page,
			Size:          size,
		})
		if err != nil {
			return nil, fmt.Errorf("list snapshots for volume %s page %d: %w", blockVolumeID, page, err)
		}

		allSnapshots = append(allSnapshots, result.Items...)

		if page >= result.TotalPages {
			break
		}
		page++
	}

	return allSnapshots, nil
}
