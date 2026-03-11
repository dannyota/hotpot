package forwardingrule

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for regional forwarding rules.
type Client struct {
	forwardingRulesClient *compute.ForwardingRulesClient
}

// NewClient creates a new GCP Compute forwarding rule client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	forwardingRulesClient, err := compute.NewForwardingRulesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create forwarding rules client: %w", err)
	}

	return &Client{
		forwardingRulesClient: forwardingRulesClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.forwardingRulesClient != nil {
		return c.forwardingRulesClient.Close()
	}
	return nil
}

// ListForwardingRules lists all regional forwarding rules in a project using aggregated list.
// Returns forwarding rules from all regions.
func (c *Client) ListForwardingRules(ctx context.Context, projectID string) ([]*computepb.ForwardingRule, error) {
	req := &computepb.AggregatedListForwardingRulesRequest{
		Project: projectID,
	}

	var forwardingRules []*computepb.ForwardingRule
	it := c.forwardingRulesClient.AggregatedList(ctx, req)

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list forwarding rules in project %s: %w", projectID, err)
		}

		forwardingRules = append(forwardingRules, pair.Value.ForwardingRules...)
	}

	return forwardingRules, nil
}
