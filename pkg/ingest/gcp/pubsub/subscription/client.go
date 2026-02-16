package subscription

import (
	"context"
	"fmt"

	pubsub "cloud.google.com/go/pubsub/apiv1"
	"cloud.google.com/go/pubsub/apiv1/pubsubpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the GCP Pub/Sub API for subscriptions.
type Client struct {
	subscriberClient *pubsub.SubscriberClient
}

// NewClient creates a new Pub/Sub subscription client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	subscriberClient, err := pubsub.NewSubscriberClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscriber client: %w", err)
	}
	return &Client{subscriberClient: subscriberClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.subscriberClient != nil {
		return c.subscriberClient.Close()
	}
	return nil
}

// ListSubscriptions lists all subscriptions in a project.
func (c *Client) ListSubscriptions(ctx context.Context, projectID string) ([]*pubsubpb.Subscription, error) {
	req := &pubsubpb.ListSubscriptionsRequest{
		Project: "projects/" + projectID,
	}

	var subscriptions []*pubsubpb.Subscription
	it := c.subscriberClient.ListSubscriptions(ctx, req)
	for {
		s, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list subscriptions in project %s: %w", projectID, err)
		}
		subscriptions = append(subscriptions, s)
	}
	return subscriptions, nil
}
