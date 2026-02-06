package globalforwardingrule

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for global forwarding rules.
type Client struct {
	globalForwardingRulesClient *compute.GlobalForwardingRulesClient
}

// NewClient creates a new GCP Compute global forwarding rule client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	globalForwardingRulesClient, err := compute.NewGlobalForwardingRulesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create global forwarding rules client: %w", err)
	}

	return &Client{
		globalForwardingRulesClient: globalForwardingRulesClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.globalForwardingRulesClient != nil {
		return c.globalForwardingRulesClient.Close()
	}
	return nil
}

// ListGlobalForwardingRules lists all global forwarding rules in a project.
func (c *Client) ListGlobalForwardingRules(ctx context.Context, projectID string) ([]*computepb.ForwardingRule, error) {
	req := &computepb.ListGlobalForwardingRulesRequest{
		Project: projectID,
	}

	var forwardingRules []*computepb.ForwardingRule
	it := c.globalForwardingRulesClient.List(ctx, req)

	for {
		fr, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list global forwarding rules in project %s: %w", projectID, err)
		}

		forwardingRules = append(forwardingRules, fr)
	}

	return forwardingRules, nil
}
