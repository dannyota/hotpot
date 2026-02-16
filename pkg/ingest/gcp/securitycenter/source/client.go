package source

import (
	"context"
	"fmt"

	securitycenter "cloud.google.com/go/securitycenter/apiv1"
	"cloud.google.com/go/securitycenter/apiv1/securitycenterpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// SourceRaw holds raw API data for an SCC source.
type SourceRaw struct {
	OrgName string
	Source  *securitycenterpb.Source
}

// Client wraps the GCP Security Command Center API for sources.
type Client struct {
	sccClient *securitycenter.Client
	entClient *ent.Client
}

// NewClient creates a new SCC source client.
func NewClient(ctx context.Context, entClient *ent.Client, opts ...option.ClientOption) (*Client, error) {
	sccClient, err := securitycenter.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create security center client: %w", err)
	}
	return &Client{sccClient: sccClient, entClient: entClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.sccClient != nil {
		return c.sccClient.Close()
	}
	return nil
}

// ListSources queries organizations from the database, then fetches sources for each.
func (c *Client) ListSources(ctx context.Context) ([]SourceRaw, error) {
	// Query organizations from database
	orgs, err := c.entClient.BronzeGCPOrganization.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query organizations from database: %w", err)
	}

	var sources []SourceRaw
	for _, org := range orgs {
		orgSources, err := c.listSourcesForOrg(ctx, org.ID)
		if err != nil {
			// Skip individual organization failures
			continue
		}
		for _, s := range orgSources {
			sources = append(sources, SourceRaw{
				OrgName: org.ID,
				Source:  s,
			})
		}
	}
	return sources, nil
}

// listSourcesForOrg fetches all sources for a single organization.
func (c *Client) listSourcesForOrg(ctx context.Context, orgName string) ([]*securitycenterpb.Source, error) {
	req := &securitycenterpb.ListSourcesRequest{
		Parent: orgName,
	}

	var sources []*securitycenterpb.Source
	it := c.sccClient.ListSources(ctx, req)
	for {
		s, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list sources for %s: %w", orgName, err)
		}
		sources = append(sources, s)
	}
	return sources, nil
}
