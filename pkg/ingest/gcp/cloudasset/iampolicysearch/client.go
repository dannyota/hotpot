package iampolicysearch

import (
	"context"
	"fmt"

	gcpasset "cloud.google.com/go/asset/apiv1"
	"cloud.google.com/go/asset/apiv1/assetpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// IAMPolicySearchRaw holds raw API data for an IAM policy search result.
type IAMPolicySearchRaw struct {
	OrgName string
	Policy  *assetpb.IamPolicySearchResult
}

// Client wraps the GCP Cloud Asset Inventory API for IAM policy search.
type Client struct {
	assetClient *gcpasset.Client
	entClient   *ent.Client
}

// NewClient creates a new Cloud Asset IAM policy search client.
func NewClient(ctx context.Context, entClient *ent.Client, opts ...option.ClientOption) (*Client, error) {
	assetClient, err := gcpasset.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud asset client: %w", err)
	}
	return &Client{assetClient: assetClient, entClient: entClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.assetClient != nil {
		return c.assetClient.Close()
	}
	return nil
}

// SearchAllIamPolicies queries organizations from the database, then searches IAM policies for each.
func (c *Client) SearchAllIamPolicies(ctx context.Context) ([]IAMPolicySearchRaw, error) {
	// Query organizations from database
	orgs, err := c.entClient.BronzeGCPOrganization.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query organizations from database: %w", err)
	}

	var policies []IAMPolicySearchRaw
	for _, org := range orgs {
		orgPolicies, err := c.searchIamPoliciesForOrg(ctx, org.ID)
		if err != nil {
			// Skip individual organization failures
			continue
		}
		for _, p := range orgPolicies {
			policies = append(policies, IAMPolicySearchRaw{
				OrgName: org.ID,
				Policy:  p,
			})
		}
	}
	return policies, nil
}

// searchIamPoliciesForOrg searches all IAM policies for a single organization.
func (c *Client) searchIamPoliciesForOrg(ctx context.Context, orgName string) ([]*assetpb.IamPolicySearchResult, error) {
	req := &assetpb.SearchAllIamPoliciesRequest{
		Scope: "organizations/" + orgName,
	}

	var policies []*assetpb.IamPolicySearchResult
	it := c.assetClient.SearchAllIamPolicies(ctx, req)
	for {
		p, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to search IAM policies for %s: %w", orgName, err)
		}
		policies = append(policies, p)
	}
	return policies, nil
}
