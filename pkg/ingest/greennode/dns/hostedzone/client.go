package hostedzone

import (
	"context"
	"fmt"

	"danny.vn/greennode"
	"danny.vn/greennode/auth"
	"danny.vn/greennode/option"
	dnsv1 "danny.vn/greennode/services/dns/v1"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
)

// Client wraps the GreenNode SDK for DNS hosted zone operations.
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

// ListHostedZones lists all DNS hosted zones.
func (c *Client) ListHostedZones(ctx context.Context) ([]*dnsv1.HostedZone, error) {
	result, err := c.sdk.DNS.ListHostedZones(ctx, &dnsv1.ListHostedZonesRequest{})
	if err != nil {
		return nil, fmt.Errorf("list hosted zones: %w", err)
	}
	return result.ListData, nil
}

// ListRecordsByHostedZoneID lists all DNS records for a hosted zone.
func (c *Client) ListRecordsByHostedZoneID(ctx context.Context, hostedZoneID string) ([]*dnsv1.DnsRecord, error) {
	result, err := c.sdk.DNS.ListRecords(ctx, &dnsv1.ListRecordsRequest{HostedZoneID: hostedZoneID})
	if err != nil {
		return nil, fmt.Errorf("list records for zone %s: %w", hostedZoneID, err)
	}
	return result.ListData, nil
}
