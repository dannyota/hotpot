package policy

import (
	"context"
	"fmt"

	binaryauthorization "cloud.google.com/go/binaryauthorization/apiv1"
	"cloud.google.com/go/binaryauthorization/apiv1/binaryauthorizationpb"
	"google.golang.org/api/option"
)

// Client wraps the GCP Binary Authorization API for policies.
type Client struct {
	binauthzClient *binaryauthorization.BinauthzManagementClient
}

// NewClient creates a new Binary Authorization policy client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	c, err := binaryauthorization.NewBinauthzManagementClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create binary authorization client: %w", err)
	}
	return &Client{binauthzClient: c}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.binauthzClient != nil {
		return c.binauthzClient.Close()
	}
	return nil
}

// GetPolicy fetches the Binary Authorization policy for a project.
// There is exactly one policy per project.
func (c *Client) GetPolicy(ctx context.Context, projectID string) (*binaryauthorizationpb.Policy, error) {
	req := &binaryauthorizationpb.GetPolicyRequest{
		Name: "projects/" + projectID + "/policy",
	}

	policy, err := c.binauthzClient.GetPolicy(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy for project %s: %w", projectID, err)
	}

	return policy, nil
}
