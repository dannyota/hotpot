package customconstraint

import (
	"context"
	"fmt"

	orgpolicy "cloud.google.com/go/orgpolicy/apiv2"
	"cloud.google.com/go/orgpolicy/apiv2/orgpolicypb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

type CustomConstraintRaw struct {
	OrgName          string
	CustomConstraint *orgpolicypb.CustomConstraint
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

func (c *Client) ListCustomConstraints(ctx context.Context) ([]CustomConstraintRaw, error) {
	orgs, err := c.entClient.BronzeGCPOrganization.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query organizations from database: %w", err)
	}

	var customConstraints []CustomConstraintRaw
	for _, org := range orgs {
		orgCustomConstraints, err := c.listCustomConstraintsForOrg(ctx, org.ID)
		if err != nil {
			continue
		}
		for _, cc := range orgCustomConstraints {
			customConstraints = append(customConstraints, CustomConstraintRaw{
				OrgName:          org.ID,
				CustomConstraint: cc,
			})
		}
	}
	return customConstraints, nil
}

func (c *Client) listCustomConstraintsForOrg(ctx context.Context, orgName string) ([]*orgpolicypb.CustomConstraint, error) {
	req := &orgpolicypb.ListCustomConstraintsRequest{
		Parent: orgName,
	}

	var customConstraints []*orgpolicypb.CustomConstraint
	it := c.orgPolicyClient.ListCustomConstraints(ctx, req)
	for {
		cc, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list custom constraints for %s: %w", orgName, err)
		}
		customConstraints = append(customConstraints, cc)
	}
	return customConstraints, nil
}
