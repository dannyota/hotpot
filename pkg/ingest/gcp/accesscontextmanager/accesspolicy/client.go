package accesspolicy

import (
	"context"
	"encoding/json"
	"fmt"

	accesscontextmanager "cloud.google.com/go/accesscontextmanager/apiv1"
	"cloud.google.com/go/accesscontextmanager/apiv1/accesscontextmanagerpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// AccessPolicyRaw holds raw API data for an access policy.
type AccessPolicyRaw struct {
	OrgName      string
	AccessPolicy *accesscontextmanagerpb.AccessPolicy
}

// Client wraps the GCP Access Context Manager API for access policies.
type Client struct {
	acmClient *accesscontextmanager.Client
	entClient *ent.Client
}

// NewClient creates a new access policy client.
func NewClient(ctx context.Context, entClient *ent.Client, opts ...option.ClientOption) (*Client, error) {
	acmClient, err := accesscontextmanager.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create access context manager client: %w", err)
	}
	return &Client{acmClient: acmClient, entClient: entClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.acmClient != nil {
		return c.acmClient.Close()
	}
	return nil
}

// ListAccessPolicies queries organizations from the database, then fetches access policies for each.
func (c *Client) ListAccessPolicies(ctx context.Context) ([]AccessPolicyRaw, error) {
	// Query organizations from database
	orgs, err := c.entClient.BronzeGCPOrganization.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query organizations from database: %w", err)
	}

	var policies []AccessPolicyRaw
	for _, org := range orgs {
		orgPolicies, err := c.listAccessPoliciesForOrg(ctx, org.ID)
		if err != nil {
			// Skip individual organization failures
			continue
		}
		for _, p := range orgPolicies {
			policies = append(policies, AccessPolicyRaw{
				OrgName:      org.ID,
				AccessPolicy: p,
			})
		}
	}
	return policies, nil
}

// listAccessPoliciesForOrg fetches all access policies for a single organization.
func (c *Client) listAccessPoliciesForOrg(ctx context.Context, orgName string) ([]*accesscontextmanagerpb.AccessPolicy, error) {
	req := &accesscontextmanagerpb.ListAccessPoliciesRequest{
		Parent: "organizations/" + orgName,
	}

	var policies []*accesscontextmanagerpb.AccessPolicy
	it := c.acmClient.ListAccessPolicies(ctx, req)
	for {
		p, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list access policies for %s: %w", orgName, err)
		}
		policies = append(policies, p)
	}
	return policies, nil
}

// scopesToJSON converts a string slice to JSON.
func scopesToJSON(scopes []string) json.RawMessage {
	if len(scopes) == 0 {
		return nil
	}
	data, err := json.Marshal(scopes)
	if err != nil {
		return nil
	}
	return data
}
