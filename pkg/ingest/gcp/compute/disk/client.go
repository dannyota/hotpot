package disk

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for disks.
type Client struct {
	disksClient *compute.DisksClient
}

// NewClient creates a new GCP Compute disk client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	disksClient, err := compute.NewDisksRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create disks client: %w", err)
	}

	return &Client{
		disksClient: disksClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.disksClient != nil {
		return c.disksClient.Close()
	}
	return nil
}

// ListDisks lists all disks in a project using aggregated list.
// Returns disks from all zones.
func (c *Client) ListDisks(ctx context.Context, projectID string) ([]*computepb.Disk, error) {
	req := &computepb.AggregatedListDisksRequest{
		Project: projectID,
	}

	var disks []*computepb.Disk
	it := c.disksClient.AggregatedList(ctx, req)

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list disks in project %s: %w", projectID, err)
		}

		disks = append(disks, pair.Value.Disks...)
	}

	return disks, nil
}

// ListDisksInZone lists disks in a specific zone.
func (c *Client) ListDisksInZone(ctx context.Context, projectID, zone string) ([]*computepb.Disk, error) {
	req := &computepb.ListDisksRequest{
		Project: projectID,
		Zone:    zone,
	}

	var disks []*computepb.Disk
	it := c.disksClient.List(ctx, req)

	for {
		disk, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list disks in zone %s: %w", zone, err)
		}

		disks = append(disks, disk)
	}

	return disks, nil
}
