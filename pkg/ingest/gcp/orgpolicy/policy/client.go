package policy

import (
	"context"
	"fmt"

	orgpolicy "cloud.google.com/go/orgpolicy/apiv2"
	"cloud.google.com/go/orgpolicy/apiv2/orgpolicypb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

type PolicyRaw struct {
	OrgName string
	Policy  *orgpolicypb.Policy
}

type Client struct {
	orgPolicyClient *orgpolicy.Client
	entClient       *ent.Client
}

func NewClient(ctx context.Context, entClient *ent.Client, opts ...option.ClientOption) (*Client, error) {
	c, err := orgpolicy.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create org policy client: %w", err)
	}
	return &Client{orgPolicyClient: c, entClient: entClient}, nil
}

func (c *Client) Close() error {
	if c.orgPolicyClient != nil {
		return c.orgPolicyClient.Close()
	}
	return nil
}

func (c *Client) ListPolicies(ctx context.Context) ([]PolicyRaw, error) {
	orgs, err := c.entClient.BronzeGCPOrganization.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query organizations from database: %w", err)
	}

	var policies []PolicyRaw
	for _, org := range orgs {
		orgPolicies, err := c.listPoliciesForOrg(ctx, org.ID)
		if err != nil {
			continue
		}
		for _, p := range orgPolicies {
			policies = append(policies, PolicyRaw{
				OrgName: org.ID,
				Policy:  p,
			})
		}
	}
	return policies, nil
}

func (c *Client) listPoliciesForOrg(ctx context.Context, orgName string) ([]*orgpolicypb.Policy, error) {
	req := &orgpolicypb.ListPoliciesRequest{
		Parent: orgName,
	}

	var policies []*orgpolicypb.Policy
	it := c.orgPolicyClient.ListPolicies(ctx, req)
	for {
		p, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list policies for %s: %w", orgName, err)
		}
		policies = append(policies, p)
	}
	return policies, nil
}
