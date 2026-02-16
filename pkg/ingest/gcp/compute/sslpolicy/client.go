package sslpolicy

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for SSL policies.
type Client struct {
	sslPoliciesClient *compute.SslPoliciesClient
}

// NewClient creates a new GCP Compute SSL policy client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	sslPoliciesClient, err := compute.NewSslPoliciesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSL policies client: %w", err)
	}

	return &Client{
		sslPoliciesClient: sslPoliciesClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.sslPoliciesClient != nil {
		return c.sslPoliciesClient.Close()
	}
	return nil
}

// ListSslPolicies lists all SSL policies in a project (global resource).
func (c *Client) ListSslPolicies(ctx context.Context, projectID string) ([]*computepb.SslPolicy, error) {
	req := &computepb.ListSslPoliciesRequest{
		Project: projectID,
	}

	var policies []*computepb.SslPolicy
	it := c.sslPoliciesClient.List(ctx, req)

	for {
		policy, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list SSL policies in project %s: %w", projectID, err)
		}

		policies = append(policies, policy)
	}

	return policies, nil
}
