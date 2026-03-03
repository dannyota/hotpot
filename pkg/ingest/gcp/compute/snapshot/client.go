package snapshot

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for snapshots.
type Client struct {
	snapshotsClient *compute.SnapshotsClient
}

// NewClient creates a new GCP Compute snapshot client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	snapshotsClient, err := compute.NewSnapshotsRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshots client: %w", err)
	}

	return &Client{
		snapshotsClient: snapshotsClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.snapshotsClient != nil {
		return c.snapshotsClient.Close()
	}
	return nil
}

// ListSnapshotsPage fetches a single page of snapshots from GCP.
func (c *Client) ListSnapshotsPage(ctx context.Context, projectID string, pageSize int, pageToken string) ([]*computepb.Snapshot, string, error) {
	it := c.snapshotsClient.List(ctx, &computepb.ListSnapshotsRequest{
		Project: projectID,
	})
	p := iterator.NewPager(it, pageSize, pageToken)

	var snapshots []*computepb.Snapshot
	nextToken, err := p.NextPage(&snapshots)
	if err != nil {
		return nil, "", fmt.Errorf("list snapshots page: %w", err)
	}

	return snapshots, nextToken, nil
}
