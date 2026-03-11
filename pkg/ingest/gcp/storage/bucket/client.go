package bucket

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	storagev1 "google.golang.org/api/storage/v1"
)

// Client wraps the GCP Cloud Storage API for buckets.
type Client struct {
	service *storagev1.Service
}

// NewClient creates a new GCP Cloud Storage client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	svc, err := storagev1.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage service: %w", err)
	}

	return &Client{service: svc}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	return nil
}

// ListBuckets lists all buckets in a project.
func (c *Client) ListBuckets(ctx context.Context, projectID string) ([]*storagev1.Bucket, error) {
	var buckets []*storagev1.Bucket

	call := c.service.Buckets.List(projectID)
	err := call.Pages(ctx, func(resp *storagev1.Buckets) error {
		buckets = append(buckets, resp.Items...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets in project %s: %w", projectID, err)
	}

	return buckets, nil
}
