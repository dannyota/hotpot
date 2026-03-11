package logbucket

import (
	"context"
	"fmt"

	logging "cloud.google.com/go/logging/apiv2"
	"cloud.google.com/go/logging/apiv2/loggingpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Cloud Logging API for buckets.
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

// ListBuckets lists all log buckets in a project across all locations.
func (c *Client) ListBuckets(ctx context.Context, projectID string) ([]*loggingpb.LogBucket, error) {
	req := &loggingpb.ListBucketsRequest{
		Parent: "projects/" + projectID + "/locations/-",
	}

	var buckets []*loggingpb.LogBucket
	it := c.configClient.ListBuckets(ctx, req)
	for {
		bucket, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list buckets in project %s: %w", projectID, err)
		}
		buckets = append(buckets, bucket)
	}
	return buckets, nil
}
