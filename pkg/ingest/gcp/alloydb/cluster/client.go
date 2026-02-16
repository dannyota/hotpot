package cluster

import (
	"context"
	"fmt"

	alloydb "cloud.google.com/go/alloydb/apiv1"
	"cloud.google.com/go/alloydb/apiv1/alloydbpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the GCP AlloyDB Admin API for clusters.
type Client struct {
	alloydbClient *alloydb.AlloyDBAdminClient
}

// NewClient creates a new GCP AlloyDB Admin client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	alloydbClient, err := alloydb.NewAlloyDBAdminClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create AlloyDB Admin client: %w", err)
	}

	return &Client{alloydbClient: alloydbClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.alloydbClient != nil {
		return c.alloydbClient.Close()
	}
	return nil
}

// ListClusters lists all AlloyDB clusters in a project across all locations.
func (c *Client) ListClusters(ctx context.Context, projectID string) ([]*alloydbpb.Cluster, error) {
	var clusters []*alloydbpb.Cluster

	parent := "projects/" + projectID + "/locations/-"
	req := &alloydbpb.ListClustersRequest{Parent: parent}

	it := c.alloydbClient.ListClusters(ctx, req)
	for {
		cluster, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list AlloyDB clusters in project %s: %w", projectID, err)
		}
		clusters = append(clusters, cluster)
	}

	return clusters, nil
}
