package iampolicy

import (
	"context"
	"fmt"

	iap "cloud.google.com/go/iap/apiv1"
	iamv1 "google.golang.org/genproto/googleapis/iam/v1"
	"google.golang.org/api/option"
)

// Client wraps the GCP Identity-Aware Proxy API for IAM policies.
type Client struct {
	iapClient *iap.IdentityAwareProxyAdminClient
}

// NewClient creates a new IAP IAM policy client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	iapClient, err := iap.NewIdentityAwareProxyAdminClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create IAP admin client: %w", err)
	}
	return &Client{iapClient: iapClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.iapClient != nil {
		return c.iapClient.Close()
	}
	return nil
}

// GetIAMPolicy fetches the IAM policy for IAP in a project.
func (c *Client) GetIAMPolicy(ctx context.Context, projectID string) (*iamv1.Policy, error) {
	req := &iamv1.GetIamPolicyRequest{
		Resource: "projects/" + projectID + "/iap_web",
	}

	policy, err := c.iapClient.GetIamPolicy(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get IAP IAM policy for project %s: %w", projectID, err)
	}

	return policy, nil
}
