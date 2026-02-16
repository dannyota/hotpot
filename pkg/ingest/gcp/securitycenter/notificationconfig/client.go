package notificationconfig

import (
	"context"
	"encoding/json"
	"fmt"

	securitycenter "cloud.google.com/go/securitycenter/apiv1"
	"cloud.google.com/go/securitycenter/apiv1/securitycenterpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// NotificationConfigRaw holds raw API data for an SCC notification config.
type NotificationConfigRaw struct {
	OrgName            string
	NotificationConfig *securitycenterpb.NotificationConfig
}

// Client wraps the GCP Security Command Center API for notification configs.
type Client struct {
	sccClient *securitycenter.Client
	entClient *ent.Client
}

// NewClient creates a new SCC notification config client.
func NewClient(ctx context.Context, entClient *ent.Client, opts ...option.ClientOption) (*Client, error) {
	sccClient, err := securitycenter.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create security center client: %w", err)
	}
	return &Client{sccClient: sccClient, entClient: entClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.sccClient != nil {
		return c.sccClient.Close()
	}
	return nil
}

// ListNotificationConfigs queries organizations from the database, then fetches notification configs for each.
func (c *Client) ListNotificationConfigs(ctx context.Context) ([]NotificationConfigRaw, error) {
	// Query organizations from database
	orgs, err := c.entClient.BronzeGCPOrganization.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query organizations from database: %w", err)
	}

	var configs []NotificationConfigRaw
	for _, org := range orgs {
		orgConfigs, err := c.listNotificationConfigsForOrg(ctx, org.ID)
		if err != nil {
			// Skip individual organization failures
			continue
		}
		for _, nc := range orgConfigs {
			configs = append(configs, NotificationConfigRaw{
				OrgName:            org.ID,
				NotificationConfig: nc,
			})
		}
	}
	return configs, nil
}

// listNotificationConfigsForOrg fetches all notification configs for a single organization.
func (c *Client) listNotificationConfigsForOrg(ctx context.Context, orgName string) ([]*securitycenterpb.NotificationConfig, error) {
	req := &securitycenterpb.ListNotificationConfigsRequest{
		Parent: orgName,
	}

	var configs []*securitycenterpb.NotificationConfig
	it := c.sccClient.ListNotificationConfigs(ctx, req)
	for {
		nc, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list notification configs for %s: %w", orgName, err)
		}
		configs = append(configs, nc)
	}
	return configs, nil
}

// streamingConfigToJSON converts a StreamingConfig to a JSON string.
func streamingConfigToJSON(nc *securitycenterpb.NotificationConfig) string {
	sc := nc.GetStreamingConfig()
	if sc == nil {
		return ""
	}
	data, err := json.Marshal(sc)
	if err != nil {
		return ""
	}
	return string(data)
}
