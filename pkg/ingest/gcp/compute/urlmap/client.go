package urlmap

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for URL maps.
type Client struct {
	urlMapsClient *compute.UrlMapsClient
}

// NewClient creates a new GCP Compute URL maps client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	umClient, err := compute.NewUrlMapsRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create URL maps client: %w", err)
	}

	return &Client{
		urlMapsClient: umClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.urlMapsClient != nil {
		return c.urlMapsClient.Close()
	}
	return nil
}

// ListUrlMaps lists all URL maps in a project using aggregated list.
func (c *Client) ListUrlMaps(ctx context.Context, projectID string) ([]*computepb.UrlMap, error) {
	req := &computepb.AggregatedListUrlMapsRequest{
		Project: projectID,
	}

	var urlMaps []*computepb.UrlMap
	it := c.urlMapsClient.AggregatedList(ctx, req)

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list URL maps in project %s: %w", projectID, err)
		}

		urlMaps = append(urlMaps, pair.Value.UrlMaps...)
	}

	return urlMaps, nil
}
