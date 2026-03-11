package uptimecheck

import (
	"context"
	"fmt"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the GCP Monitoring API for uptime check configs.
type Client struct {
	uptimeCheckClient *monitoring.UptimeCheckClient
}

// NewClient creates a new Monitoring uptime check client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	c, err := monitoring.NewUptimeCheckClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create uptime check client: %w", err)
	}
	return &Client{uptimeCheckClient: c}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.uptimeCheckClient != nil {
		return c.uptimeCheckClient.Close()
	}
	return nil
}

// ListUptimeCheckConfigs lists all uptime check configs in a project.
func (c *Client) ListUptimeCheckConfigs(ctx context.Context, projectID string) ([]*monitoringpb.UptimeCheckConfig, error) {
	req := &monitoringpb.ListUptimeCheckConfigsRequest{
		Parent: "projects/" + projectID,
	}

	var configs []*monitoringpb.UptimeCheckConfig
	it := c.uptimeCheckClient.ListUptimeCheckConfigs(ctx, req)
	for {
		cfg, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list uptime check configs for project %s: %w", projectID, err)
		}
		configs = append(configs, cfg)
	}
	return configs, nil
}
