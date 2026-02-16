package serviceperimeter

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

// ServicePerimeterRaw holds raw API data for a service perimeter.
type ServicePerimeterRaw struct {
	OrgName          string
	AccessPolicyName string
	ServicePerimeter *accesscontextmanagerpb.ServicePerimeter
}

// Client wraps the GCP Access Context Manager API for service perimeters.
type Client struct {
	acmClient *accesscontextmanager.Client
	entClient *ent.Client
}

// NewClient creates a new service perimeter client.
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

// ListServicePerimeters queries access policies from the database, then fetches service perimeters for each.
func (c *Client) ListServicePerimeters(ctx context.Context) ([]ServicePerimeterRaw, error) {
	// Query access policies from database
	policies, err := c.entClient.BronzeGCPAccessContextManagerAccessPolicy.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query access policies from database: %w", err)
	}

	var perimeters []ServicePerimeterRaw
	for _, policy := range policies {
		policyPerimeters, err := c.listServicePerimetersForPolicy(ctx, policy.ID)
		if err != nil {
			// Skip individual policy failures
			continue
		}
		for _, p := range policyPerimeters {
			perimeters = append(perimeters, ServicePerimeterRaw{
				OrgName:          policy.OrganizationID,
				AccessPolicyName: policy.ID,
				ServicePerimeter: p,
			})
		}
	}
	return perimeters, nil
}

// listServicePerimetersForPolicy fetches all service perimeters for a single access policy.
func (c *Client) listServicePerimetersForPolicy(ctx context.Context, policyName string) ([]*accesscontextmanagerpb.ServicePerimeter, error) {
	req := &accesscontextmanagerpb.ListServicePerimetersRequest{
		Parent: policyName,
	}

	var perimeters []*accesscontextmanagerpb.ServicePerimeter
	it := c.acmClient.ListServicePerimeters(ctx, req)
	for {
		p, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list service perimeters for %s: %w", policyName, err)
		}
		perimeters = append(perimeters, p)
	}
	return perimeters, nil
}

// servicePerimeterConfigToJSON converts a ServicePerimeterConfig proto to JSON.
func servicePerimeterConfigToJSON(config *accesscontextmanagerpb.ServicePerimeterConfig) json.RawMessage {
	if config == nil {
		return nil
	}
	data, err := protojson.Marshal(config)
	if err != nil {
		return nil
	}
	return data
}
