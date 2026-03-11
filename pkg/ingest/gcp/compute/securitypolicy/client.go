package securitypolicy

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for security policies.
type Client struct {
	securityPoliciesClient *compute.SecurityPoliciesClient
}

// NewClient creates a new GCP Compute security policy client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	securityPoliciesClient, err := compute.NewSecurityPoliciesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create security policies client: %w", err)
	}

	return &Client{
		securityPoliciesClient: securityPoliciesClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.securityPoliciesClient != nil {
		return c.securityPoliciesClient.Close()
	}
	return nil
}

// ListSecurityPolicies lists all security policies in a project (global resource).
func (c *Client) ListSecurityPolicies(ctx context.Context, projectID string) ([]*computepb.SecurityPolicy, error) {
	req := &computepb.ListSecurityPoliciesRequest{
		Project: projectID,
	}

	var policies []*computepb.SecurityPolicy
	it := c.securityPoliciesClient.List(ctx, req)

	for {
		policy, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list security policies in project %s: %w", projectID, err)
		}

		policies = append(policies, policy)
	}

	return policies, nil
}
