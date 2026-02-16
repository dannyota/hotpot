package orgiampolicy

import (
	"context"
	"fmt"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/iam/apiv1/iampb"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// OrgIamPolicyRaw holds raw API data for an organization IAM policy.
type OrgIamPolicyRaw struct {
	OrgName string
	Policy  *iampb.Policy
}

// Client wraps the GCP Resource Manager API for organization IAM policies.
type Client struct {
	orgsClient *resourcemanager.OrganizationsClient
	entClient  *ent.Client
}

// NewClient creates a new GCP Resource Manager organization IAM policy client.
func NewClient(ctx context.Context, entClient *ent.Client, opts ...option.ClientOption) (*Client, error) {
	orgsClient, err := resourcemanager.NewOrganizationsClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create organizations client: %w", err)
	}
	return &Client{orgsClient: orgsClient, entClient: entClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.orgsClient != nil {
		return c.orgsClient.Close()
	}
	return nil
}

// ListOrgIamPolicies queries organizations from the database, then fetches IAM policies for each.
func (c *Client) ListOrgIamPolicies(ctx context.Context) ([]OrgIamPolicyRaw, error) {
	// Query organizations from database
	orgs, err := c.entClient.BronzeGCPOrganization.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query organizations from database: %w", err)
	}

	var policies []OrgIamPolicyRaw
	for _, org := range orgs {
		policy, err := c.orgsClient.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
			Resource: org.ID,
		})
		if err != nil {
			// Skip individual organization failures
			continue
		}
		policies = append(policies, OrgIamPolicyRaw{
			OrgName: org.ID,
			Policy:  policy,
		})
	}
	return policies, nil
}
