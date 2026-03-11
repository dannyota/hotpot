package connector

import (
	"context"
	"fmt"

	vpcaccess "cloud.google.com/go/vpcaccess/apiv1"
	"cloud.google.com/go/vpcaccess/apiv1/vpcaccesspb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the GCP VPC Access API for connectors.
type Client struct {
	vpcaccessClient *vpcaccess.Client
}

// NewClient creates a new GCP VPC Access connector client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	vpcClient, err := vpcaccess.NewRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create vpc access client: %w", err)
	}

	return &Client{
		vpcaccessClient: vpcClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.vpcaccessClient != nil {
		return c.vpcaccessClient.Close()
	}
	return nil
}

// ListConnectors lists all VPC Access connectors across the given regions.
// The VPC Access API does not support wildcard locations, so regions must be provided.
func (c *Client) ListConnectors(ctx context.Context, projectID string, regions []string) ([]*vpcaccesspb.Connector, error) {
	var connectors []*vpcaccesspb.Connector

	for _, region := range regions {
		parent := fmt.Sprintf("projects/%s/locations/%s", projectID, region)
		req := &vpcaccesspb.ListConnectorsRequest{Parent: parent}

		it := c.vpcaccessClient.ListConnectors(ctx, req)
		for {
			connector, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				// Skip regions that return errors (API not enabled or region doesn't support VPC Access)
				break
			}

			connectors = append(connectors, connector)
		}
	}

	return connectors, nil
}
