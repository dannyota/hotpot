package settings

import (
	"context"
	"fmt"

	iap "cloud.google.com/go/iap/apiv1"
	"cloud.google.com/go/iap/apiv1/iappb"
	"google.golang.org/api/option"
)

// Client wraps the GCP Identity-Aware Proxy API for settings.
type Client struct {
	iapClient *iap.IdentityAwareProxyAdminClient
}

// NewClient creates a new IAP settings client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	iapClient, err := iap.NewIdentityAwareProxyAdminClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create IAP admin client: %w", err)
	}
	return &Client{iapClient: iapClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.iapClient != nil {
		return c.iapClient.Close()
	}
	return nil
}

// GetSettings fetches IAP settings for a project.
func (c *Client) GetSettings(ctx context.Context, projectID string) (*iappb.IapSettings, error) {
	req := &iappb.GetIapSettingsRequest{
		Name: "projects/" + projectID + "/iap_web",
	}

	settings, err := c.iapClient.GetIapSettings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get IAP settings for project %s: %w", projectID, err)
	}

	return settings, nil
}
