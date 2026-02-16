package instance

import (
	"context"
	"fmt"

	redis "cloud.google.com/go/redis/apiv1"
	"cloud.google.com/go/redis/apiv1/redispb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the GCP Memorystore Redis API for instances.
type Client struct {
	redisClient *redis.CloudRedisClient
}

// NewClient creates a new GCP Memorystore Redis client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	redisClient, err := redis.NewCloudRedisClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Memorystore Redis client: %w", err)
	}

	return &Client{redisClient: redisClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.redisClient != nil {
		return c.redisClient.Close()
	}
	return nil
}

// ListInstances lists all Memorystore Redis instances in a project across all locations.
func (c *Client) ListInstances(ctx context.Context, projectID string) ([]*redispb.Instance, error) {
	var instances []*redispb.Instance

	parent := "projects/" + projectID + "/locations/-"
	req := &redispb.ListInstancesRequest{Parent: parent}

	it := c.redisClient.ListInstances(ctx, req)
	for {
		inst, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list Redis instances in project %s: %w", projectID, err)
		}
		instances = append(instances, inst)
	}

	return instances, nil
}
