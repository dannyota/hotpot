package sshkey

import (
	"context"
	"fmt"

	"danny.vn/greennode"
	"danny.vn/greennode/auth"
	"danny.vn/greennode/option"
	computev2 "danny.vn/greennode/services/compute/v2"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
)

// Client wraps the GreenNode SDK for SSH key operations.
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

// ListSSHKeys lists all SSH keys, handling pagination.
func (c *Client) ListSSHKeys(ctx context.Context) ([]*computev2.SSHKey, error) {
	var allKeys []*computev2.SSHKey
	page := 1
	size := 50

	for {
		result, err := c.sdk.Compute.ListSSHKeys(ctx, computev2.NewListSSHKeysRequest(page, size))
		if err != nil {
			return nil, fmt.Errorf("list ssh keys page %d: %w", page, err)
		}

		allKeys = append(allKeys, result.Items...)

		if page >= result.TotalPage {
			break
		}
		page++
	}

	return allKeys, nil
}
