package packetmirroring

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for packet mirrorings.
type Client struct {
	packetMirroringsClient *compute.PacketMirroringsClient
}

// NewClient creates a new GCP Compute packet mirroring client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	packetMirroringsClient, err := compute.NewPacketMirroringsRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create packet mirrorings client: %w", err)
	}

	return &Client{
		packetMirroringsClient: packetMirroringsClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.packetMirroringsClient != nil {
		return c.packetMirroringsClient.Close()
	}
	return nil
}

// ListPacketMirrorings lists all packet mirrorings in a project using aggregated list.
// Returns packet mirrorings from all regions.
func (c *Client) ListPacketMirrorings(ctx context.Context, projectID string) ([]*computepb.PacketMirroring, error) {
	req := &computepb.AggregatedListPacketMirroringsRequest{
		Project: projectID,
	}

	var packetMirrorings []*computepb.PacketMirroring
	it := c.packetMirroringsClient.AggregatedList(ctx, req)

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list packet mirrorings in project %s: %w", projectID, err)
		}

		packetMirrorings = append(packetMirrorings, pair.Value.PacketMirrorings...)
	}

	return packetMirrorings, nil
}
