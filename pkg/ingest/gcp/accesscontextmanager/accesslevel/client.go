package accesslevel

import (
	"context"
	"encoding/json"
	"fmt"

	accesscontextmanager "cloud.google.com/go/accesscontextmanager/apiv1"
	"cloud.google.com/go/accesscontextmanager/apiv1/accesscontextmanagerpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// AccessLevelRaw holds raw API data for an access level.
type AccessLevelRaw struct {
	OrgName          string
	AccessPolicyName string
	AccessLevel      *accesscontextmanagerpb.AccessLevel
}

// Client wraps the GCP Access Context Manager API for access levels.
type Client struct {
	acmClient *accesscontextmanager.Client
	entClient *ent.Client
}

// NewClient creates a new access level client.
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

// ListAccessLevels queries access policies from the database, then fetches access levels for each.
func (c *Client) ListAccessLevels(ctx context.Context) ([]AccessLevelRaw, error) {
	// Query access policies from database
	policies, err := c.entClient.BronzeGCPAccessContextManagerAccessPolicy.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query access policies from database: %w", err)
	}

	var levels []AccessLevelRaw
	for _, policy := range policies {
		policyLevels, err := c.listAccessLevelsForPolicy(ctx, policy.ID)
		if err != nil {
			// Skip individual policy failures
			continue
		}
		for _, l := range policyLevels {
			levels = append(levels, AccessLevelRaw{
				OrgName:          policy.OrganizationID,
				AccessPolicyName: policy.ID,
				AccessLevel:      l,
			})
		}
	}
	return levels, nil
}

// listAccessLevelsForPolicy fetches all access levels for a single access policy.
func (c *Client) listAccessLevelsForPolicy(ctx context.Context, policyName string) ([]*accesscontextmanagerpb.AccessLevel, error) {
	req := &accesscontextmanagerpb.ListAccessLevelsRequest{
		Parent: policyName,
	}

	var levels []*accesscontextmanagerpb.AccessLevel
	it := c.acmClient.ListAccessLevels(ctx, req)
	for {
		l, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list access levels for %s: %w", policyName, err)
		}
		levels = append(levels, l)
	}
	return levels, nil
}

// basicLevelToJSON converts BasicLevel proto to JSON.
func basicLevelToJSON(level *accesscontextmanagerpb.BasicLevel) json.RawMessage {
	if level == nil {
		return nil
	}
	data, err := protojson.Marshal(level)
	if err != nil {
		return nil
	}
	return data
}

// customLevelToJSON converts CustomLevel proto to JSON.
func customLevelToJSON(level *accesscontextmanagerpb.CustomLevel) json.RawMessage {
	if level == nil {
		return nil
	}
	data, err := protojson.Marshal(level)
	if err != nil {
		return nil
	}
	return data
}
