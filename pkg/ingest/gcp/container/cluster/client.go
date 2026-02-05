package cluster

import (
	"context"
	"fmt"

	container "cloud.google.com/go/container/apiv1"
	"cloud.google.com/go/container/apiv1/containerpb"
	"google.golang.org/api/option"
)

// Client wraps GCP Container API for clusters.
type Client struct {
	clusterManager *container.ClusterManagerClient
}

// NewClient creates a new GCP Container cluster client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	clusterManager, err := container.NewClusterManagerClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create cluster manager client: %w", err)
	}

	return &Client{
		clusterManager: clusterManager,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.clusterManager != nil {
		return c.clusterManager.Close()
	}
	return nil
}

// ListClusters lists all clusters in a project across all locations.
func (c *Client) ListClusters(ctx context.Context, projectID string) ([]*containerpb.Cluster, error) {
	req := &containerpb.ListClustersRequest{
		Parent: fmt.Sprintf("projects/%s/locations/-", projectID),
	}

	resp, err := c.clusterManager.ListClusters(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("list clusters in project %s: %w", projectID, err)
	}

	return resp.Clusters, nil
}
