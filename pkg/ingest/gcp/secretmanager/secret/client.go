package secret

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the GCP Secret Manager API for secrets.
type Client struct {
	smClient *secretmanager.Client
}

// NewClient creates a new GCP Secret Manager client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	smClient, err := secretmanager.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Secret Manager client: %w", err)
	}

	return &Client{smClient: smClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.smClient != nil {
		return c.smClient.Close()
	}
	return nil
}

// ListSecrets lists all secrets in a project.
func (c *Client) ListSecrets(ctx context.Context, projectID string) ([]*secretmanagerpb.Secret, error) {
	var secrets []*secretmanagerpb.Secret

	parent := "projects/" + projectID
	req := &secretmanagerpb.ListSecretsRequest{Parent: parent}

	it := c.smClient.ListSecrets(ctx, req)
	for {
		s, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list secrets in project %s: %w", projectID, err)
		}
		secrets = append(secrets, s)
	}

	return secrets, nil
}
