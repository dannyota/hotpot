package cluster

import (
	"context"
	"fmt"

	dataproc "cloud.google.com/go/dataproc/v2/apiv1"
	"cloud.google.com/go/dataproc/v2/apiv1/dataprocpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the GCP Dataproc API for clusters.
type Client struct {
	dpClient *dataproc.ClusterControllerClient
}

// NewClient creates a new GCP Dataproc cluster client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	dpClient, err := dataproc.NewClusterControllerClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Dataproc cluster controller client: %w", err)
	}

	return &Client{dpClient: dpClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.dpClient != nil {
		return c.dpClient.Close()
	}
	return nil
}

// ListClusters lists all Dataproc clusters in a project across all regions.
func (c *Client) ListClusters(ctx context.Context, projectID string) ([]*dataprocpb.Cluster, error) {
	var clusters []*dataprocpb.Cluster

	req := &dataprocpb.ListClustersRequest{
		ProjectId: projectID,
		Region:    "-",
	}

	it := c.dpClient.ListClusters(ctx, req)
	for {
		cluster, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list Dataproc clusters in project %s: %w", projectID, err)
		}
		clusters = append(clusters, cluster)
	}

	return clusters, nil
}
