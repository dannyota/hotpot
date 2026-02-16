package logexclusion

import (
	"context"
	"fmt"

	logging "cloud.google.com/go/logging/apiv2"
	"cloud.google.com/go/logging/apiv2/loggingpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Cloud Logging API for log exclusions.
type Client struct {
	configClient *logging.ConfigClient
}

// NewClient creates a new GCP Cloud Logging config client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	configClient, err := logging.NewConfigClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create logging config client: %w", err)
	}
	return &Client{configClient: configClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.configClient != nil {
		return c.configClient.Close()
	}
	return nil
}

// ListExclusions lists all log exclusions in a project.
func (c *Client) ListExclusions(ctx context.Context, projectID string) ([]*loggingpb.LogExclusion, error) {
	req := &loggingpb.ListExclusionsRequest{
		Parent: "projects/" + projectID,
	}

	var exclusions []*loggingpb.LogExclusion
	it := c.configClient.ListExclusions(ctx, req)
	for {
		exclusion, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list exclusions in project %s: %w", projectID, err)
		}
		exclusions = append(exclusions, exclusion)
	}
	return exclusions, nil
}
