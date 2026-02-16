package topic

import (
	"context"
	"fmt"

	pubsub "cloud.google.com/go/pubsub/apiv1"
	"cloud.google.com/go/pubsub/apiv1/pubsubpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the GCP Pub/Sub API for topics.
type Client struct {
	publisherClient *pubsub.PublisherClient
}

// NewClient creates a new Pub/Sub topic client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	publisherClient, err := pubsub.NewPublisherClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher client: %w", err)
	}
	return &Client{publisherClient: publisherClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.publisherClient != nil {
		return c.publisherClient.Close()
	}
	return nil
}

// ListTopics lists all topics in a project.
func (c *Client) ListTopics(ctx context.Context, projectID string) ([]*pubsubpb.Topic, error) {
	req := &pubsubpb.ListTopicsRequest{
		Project: "projects/" + projectID,
	}

	var topics []*pubsubpb.Topic
	it := c.publisherClient.ListTopics(ctx, req)
	for {
		t, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list topics in project %s: %w", projectID, err)
		}
		topics = append(topics, t)
	}
	return topics, nil
}
