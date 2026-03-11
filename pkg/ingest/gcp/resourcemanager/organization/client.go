package organization

import (
	"context"
	"fmt"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Cloud Resource Manager API for organizations.
type Client struct {
	organizationsClient *resourcemanager.OrganizationsClient
}

// NewClient creates a new GCP Resource Manager organizations client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	organizationsClient, err := resourcemanager.NewOrganizationsClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create organizations client: %w", err)
	}

	return &Client{
		organizationsClient: organizationsClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.organizationsClient != nil {
		return c.organizationsClient.Close()
	}
	return nil
}

// SearchOrganizations searches for all organizations accessible by the service account.
// Returns organizations where the caller has resourcemanager.organizations.get permission.
func (c *Client) SearchOrganizations(ctx context.Context) ([]*resourcemanagerpb.Organization, error) {
	req := &resourcemanagerpb.SearchOrganizationsRequest{}

	var organizations []*resourcemanagerpb.Organization
	it := c.organizationsClient.SearchOrganizations(ctx, req)

	for {
		org, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to search organizations: %w", err)
		}

		organizations = append(organizations, org)
	}

	return organizations, nil
}
