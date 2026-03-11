package userimage

import (
	"context"
	"fmt"

	"danny.vn/gnode"
	"danny.vn/gnode/auth"
	"danny.vn/gnode/option"
	computev2 "danny.vn/gnode/services/compute/v2"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
)

// Client wraps the GreenNode SDK for user image operations.
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

// ListUserImages lists all user images, handling pagination.
func (c *Client) ListUserImages(ctx context.Context) ([]*computev2.UserImage, error) {
	var all []*computev2.UserImage
	page := 1
	size := 50

	for {
		result, err := c.sdk.Compute.ListUserImages(ctx, &computev2.ListUserImagesRequest{
			Page: page,
			Size: size,
		})
		if err != nil {
			return nil, fmt.Errorf("list user images page %d: %w", page, err)
		}

		all = append(all, result.Items...)

		if page >= result.TotalPage {
			break
		}
		page++
	}

	return all, nil
}
