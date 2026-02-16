package constraint

import (
	"context"
	"fmt"

	orgpolicy "cloud.google.com/go/orgpolicy/apiv2"
	"cloud.google.com/go/orgpolicy/apiv2/orgpolicypb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

type ConstraintRaw struct {
	OrgName    string
	Constraint *orgpolicypb.Constraint
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

func (c *Client) ListConstraints(ctx context.Context) ([]ConstraintRaw, error) {
	orgs, err := c.entClient.BronzeGCPOrganization.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query organizations from database: %w", err)
	}

	var constraints []ConstraintRaw
	for _, org := range orgs {
		orgConstraints, err := c.listConstraintsForOrg(ctx, org.ID)
		if err != nil {
			continue
		}
		for _, con := range orgConstraints {
			constraints = append(constraints, ConstraintRaw{
				OrgName:    org.ID,
				Constraint: con,
			})
		}
	}
	return constraints, nil
}

func (c *Client) listConstraintsForOrg(ctx context.Context, orgName string) ([]*orgpolicypb.Constraint, error) {
	req := &orgpolicypb.ListConstraintsRequest{
		Parent: orgName,
	}

	var constraints []*orgpolicypb.Constraint
	it := c.orgPolicyClient.ListConstraints(ctx, req)
	for {
		con, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list constraints for %s: %w", orgName, err)
		}
		constraints = append(constraints, con)
	}
	return constraints, nil
}
