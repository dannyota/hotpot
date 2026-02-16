package folderiampolicy

import (
	"context"
	"fmt"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/iam/apiv1/iampb"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// FolderIamPolicyRaw holds raw API data for a folder IAM policy.
type FolderIamPolicyRaw struct {
	FolderName string
	Policy     *iampb.Policy
}

// Client wraps the GCP Resource Manager API for folder IAM policies.
type Client struct {
	foldersClient *resourcemanager.FoldersClient
	entClient     *ent.Client
}

// NewClient creates a new GCP Resource Manager folder IAM policy client.
func NewClient(ctx context.Context, entClient *ent.Client, opts ...option.ClientOption) (*Client, error) {
	foldersClient, err := resourcemanager.NewFoldersClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create folders client: %w", err)
	}
	return &Client{foldersClient: foldersClient, entClient: entClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.foldersClient != nil {
		return c.foldersClient.Close()
	}
	return nil
}

// ListFolderIamPolicies queries folders from the database, then fetches IAM policies for each.
func (c *Client) ListFolderIamPolicies(ctx context.Context) ([]FolderIamPolicyRaw, error) {
	// Query folders from database
	folders, err := c.entClient.BronzeGCPFolder.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query folders from database: %w", err)
	}

	var policies []FolderIamPolicyRaw
	for _, folder := range folders {
		policy, err := c.foldersClient.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
			Resource: folder.ID,
		})
		if err != nil {
			// Skip individual folder failures
			continue
		}
		policies = append(policies, FolderIamPolicyRaw{
			FolderName: folder.ID,
			Policy:     policy,
		})
	}
	return policies, nil
}
