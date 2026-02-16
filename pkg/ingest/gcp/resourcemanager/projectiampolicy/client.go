package projectiampolicy

import (
	"context"
	"fmt"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/iam/apiv1/iampb"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpproject"
)

// ProjectIamPolicyRaw holds raw API data for a project IAM policy.
type ProjectIamPolicyRaw struct {
	ProjectID string
	Policy    *iampb.Policy
}

// Client wraps the GCP Resource Manager API for project IAM policies.
type Client struct {
	client    *resourcemanager.ProjectsClient
	entClient *ent.Client
}

// NewClient creates a new GCP Resource Manager project IAM policy client.
func NewClient(ctx context.Context, entClient *ent.Client, opts ...option.ClientOption) (*Client, error) {
	client, err := resourcemanager.NewProjectsClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource manager projects client: %w", err)
	}
	return &Client{client: client, entClient: entClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	return c.client.Close()
}

// GetProjectIamPolicy queries the project from the database and fetches its IAM policy.
func (c *Client) GetProjectIamPolicy(ctx context.Context, projectID string) (*ProjectIamPolicyRaw, error) {
	// Query the project from database to verify it exists
	_, err := c.entClient.BronzeGCPProject.Query().
		Where(bronzegcpproject.ID(projectID)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query project from database: %w", err)
	}

	// Fetch the IAM policy for this project
	policy, err := c.client.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: "projects/" + projectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get IAM policy for project %s: %w", projectID, err)
	}

	return &ProjectIamPolicyRaw{
		ProjectID: projectID,
		Policy:    policy,
	}, nil
}
