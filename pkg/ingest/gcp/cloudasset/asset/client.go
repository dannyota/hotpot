package asset

import (
	"context"
	"fmt"

	gcpasset "cloud.google.com/go/asset/apiv1"
	"cloud.google.com/go/asset/apiv1/assetpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// AssetRaw holds raw API data for a Cloud Asset Inventory asset.
type AssetRaw struct {
	OrgName string
	Asset   *assetpb.Asset
}

// Client wraps the GCP Cloud Asset Inventory API for assets.
type Client struct {
	assetClient *gcpasset.Client
	entClient   *ent.Client
}

// NewClient creates a new Cloud Asset Inventory asset client.
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

// ListAssets queries organizations from the database, then fetches assets for each.
func (c *Client) ListAssets(ctx context.Context) ([]AssetRaw, error) {
	// Query organizations from database
	orgs, err := c.entClient.BronzeGCPOrganization.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query organizations from database: %w", err)
	}

	var assets []AssetRaw
	for _, org := range orgs {
		orgAssets, err := c.listAssetsForOrg(ctx, org.ID)
		if err != nil {
			// Skip individual organization failures
			continue
		}
		for _, a := range orgAssets {
			assets = append(assets, AssetRaw{
				OrgName: org.ID,
				Asset:   a,
			})
		}
	}
	return assets, nil
}

// listAssetsForOrg fetches all assets for a single organization.
func (c *Client) listAssetsForOrg(ctx context.Context, orgName string) ([]*assetpb.Asset, error) {
	req := &assetpb.ListAssetsRequest{
		Parent: "organizations/" + orgName,
	}

	var assets []*assetpb.Asset
	it := c.assetClient.ListAssets(ctx, req)
	for {
		a, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list assets for %s: %w", orgName, err)
		}
		assets = append(assets, a)
	}
	return assets, nil
}
