package resourcesearch

import (
	"context"
	"fmt"

	gcpasset "cloud.google.com/go/asset/apiv1"
	"cloud.google.com/go/asset/apiv1/assetpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// ResourceSearchRaw holds raw API data for a resource search result.
type ResourceSearchRaw struct {
	OrgName  string
	Resource *assetpb.ResourceSearchResult
}

// Client wraps the GCP Cloud Asset Inventory API for resource search.
type Client struct {
	assetClient *gcpasset.Client
	entClient   *ent.Client
}

// NewClient creates a new Cloud Asset resource search client.
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

// SearchAllResources queries organizations from the database, then searches resources for each.
func (c *Client) SearchAllResources(ctx context.Context) ([]ResourceSearchRaw, error) {
	// Query organizations from database
	orgs, err := c.entClient.BronzeGCPOrganization.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query organizations from database: %w", err)
	}

	var resources []ResourceSearchRaw
	for _, org := range orgs {
		orgResources, err := c.searchResourcesForOrg(ctx, org.ID)
		if err != nil {
			// Skip individual organization failures
			continue
		}
		for _, r := range orgResources {
			resources = append(resources, ResourceSearchRaw{
				OrgName:  org.ID,
				Resource: r,
			})
		}
	}
	return resources, nil
}

// searchResourcesForOrg searches all resources for a single organization.
func (c *Client) searchResourcesForOrg(ctx context.Context, orgName string) ([]*assetpb.ResourceSearchResult, error) {
	req := &assetpb.SearchAllResourcesRequest{
		Scope: "organizations/" + orgName,
	}

	var resources []*assetpb.ResourceSearchResult
	it := c.assetClient.SearchAllResources(ctx, req)
	for {
		r, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to search resources for %s: %w", orgName, err)
		}
		resources = append(resources, r)
	}
	return resources, nil
}
