package subnet

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

// Client wraps the GreenNode SDK for subnet operations.
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

// ListSubnets lists all subnets across all networks, handling pagination for networks.
func (c *Client) ListSubnets(ctx context.Context) ([]*networkv2.Subnet, error) {
	// First list all networks (paginated)
	var allNetworks []*networkv2.Network
	page := 1
	size := 50

	for {
		result, err := c.sdk.Network.ListNetworks(ctx, &networkv2.ListNetworksRequest{
			Page: page,
			Size: size,
		})
		if err != nil {
			return nil, fmt.Errorf("list networks page %d: %w", page, err)
		}

		allNetworks = append(allNetworks, result.Items...)

		if page >= result.TotalPage {
			break
		}
		page++
	}

	// Then list subnets for each network
	var allSubnets []*networkv2.Subnet
	for _, net := range allNetworks {
		result, err := c.sdk.Network.ListSubnetsByNetworkID(ctx, &networkv2.ListSubnetsByNetworkIDRequest{
			NetworkID: net.UUID,
		})
		if err != nil {
			return nil, fmt.Errorf("list subnets for network %s: %w", net.UUID, err)
		}

		allSubnets = append(allSubnets, result.Items...)
	}

	return allSubnets, nil
}
