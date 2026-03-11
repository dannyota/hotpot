package sink

import (
	"context"
	"fmt"

	logging "cloud.google.com/go/logging/apiv2"
	"cloud.google.com/go/logging/apiv2/loggingpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Cloud Logging API for sinks.
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

// ListSinks lists all sinks in a project.
func (c *Client) ListSinks(ctx context.Context, projectID string) ([]*loggingpb.LogSink, error) {
	req := &loggingpb.ListSinksRequest{
		Parent: "projects/" + projectID,
	}

	var sinks []*loggingpb.LogSink
	it := c.configClient.ListSinks(ctx, req)
	for {
		sink, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list sinks in project %s: %w", projectID, err)
		}
		sinks = append(sinks, sink)
	}
	return sinks, nil
}
