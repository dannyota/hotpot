package cluster

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigtable"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpbigtableinstance"
)

// ClusterRaw holds raw API data for a Bigtable cluster with its parent instance.
type ClusterRaw struct {
	InstanceName string
	InstanceID   string
	Cluster      *bigtable.ClusterInfo
}

// Client wraps the GCP Bigtable Instance Admin API for clusters.
type Client struct {
	adminClient *bigtable.InstanceAdminClient
	entClient   *ent.Client
	projectID   string
}

// NewClient creates a new Bigtable cluster client.
func NewClient(ctx context.Context, projectID string, entClient *ent.Client, opts ...option.ClientOption) (*Client, error) {
	adminClient, err := bigtable.NewInstanceAdminClient(ctx, projectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create bigtable instance admin client: %w", err)
	}
	return &Client{adminClient: adminClient, entClient: entClient, projectID: projectID}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.adminClient != nil {
		return c.adminClient.Close()
	}
	return nil
}

// ListClusters queries instances from the database, then fetches clusters for each.
func (c *Client) ListClusters(ctx context.Context, projectID string) ([]ClusterRaw, error) {
	// Query instances from database
	instances, err := c.entClient.BronzeGCPBigtableInstance.Query().
		Where(bronzegcpbigtableinstance.ProjectID(projectID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query bigtable instances from database: %w", err)
	}

	var clusters []ClusterRaw
	for _, inst := range instances {
		// Extract instance ID from resource name: projects/{project}/instances/{instance}
		instanceID := extractInstanceID(inst.ID)
		if instanceID == "" {
			continue
		}

		instClusters, err := c.adminClient.Clusters(ctx, instanceID)
		if err != nil {
			// Skip individual instance failures
			continue
		}
		for _, cl := range instClusters {
			clusters = append(clusters, ClusterRaw{
				InstanceName: inst.ID,
				InstanceID:   instanceID,
				Cluster:      cl,
			})
		}
	}
	return clusters, nil
}

// extractInstanceID extracts the instance ID from a resource name.
// Input: "projects/{project}/instances/{instance}"
// Output: "{instance}"
func extractInstanceID(resourceName string) string {
	// Find last "/" and return everything after it
	for i := len(resourceName) - 1; i >= 0; i-- {
		if resourceName[i] == '/' {
			return resourceName[i+1:]
		}
	}
	return resourceName
}
