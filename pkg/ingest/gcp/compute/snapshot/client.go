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

// ListSnapshots lists all snapshots in a project (global resource).
func (c *Client) ListSnapshots(ctx context.Context, projectID string) ([]*computepb.Snapshot, error) {
	req := &computepb.ListSnapshotsRequest{
		Project: projectID,
	}

	var snapshots []*computepb.Snapshot
	it := c.snapshotsClient.List(ctx, req)

	for {
		snap, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list snapshots in project %s: %w", projectID, err)
		}

		snapshots = append(snapshots, snap)
	}

	return snapshots, nil
}
