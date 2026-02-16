package finding

import (
	"context"
	"fmt"

	securitycenter "cloud.google.com/go/securitycenter/apiv1"
	"cloud.google.com/go/securitycenter/apiv1/securitycenterpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// FindingRaw holds raw API data for an SCC finding.
type FindingRaw struct {
	OrgName    string
	SourceName string
	Finding    *securitycenterpb.ListFindingsResponse_ListFindingsResult
}

// Client wraps the GCP Security Command Center API for findings.
type Client struct {
	sccClient *securitycenter.Client
	entClient *ent.Client
}

// NewClient creates a new SCC finding client.
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

// ListFindings queries organizations from the database, fetches sources per org,
// then fetches findings per source.
func (c *Client) ListFindings(ctx context.Context) ([]FindingRaw, error) {
	// Query organizations from database
	orgs, err := c.entClient.BronzeGCPOrganization.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query organizations from database: %w", err)
	}

	var findings []FindingRaw
	for _, org := range orgs {
		// Query sources for this org from database
		sources, err := c.entClient.BronzeGCPSecurityCenterSource.Query().All(ctx)
		if err != nil {
			continue
		}

		for _, source := range sources {
			// Only process sources belonging to this org
			if source.OrganizationID != org.ID {
				continue
			}

			sourceFindingsRaw, err := c.listFindingsForSource(ctx, source.ID)
			if err != nil {
				// Skip individual source failures
				continue
			}
			for _, f := range sourceFindingsRaw {
				findings = append(findings, FindingRaw{
					OrgName:    org.ID,
					SourceName: source.ID,
					Finding:    f,
				})
			}
		}
	}
	return findings, nil
}

// listFindingsForSource fetches all findings for a single source.
func (c *Client) listFindingsForSource(ctx context.Context, sourceName string) ([]*securitycenterpb.ListFindingsResponse_ListFindingsResult, error) {
	req := &securitycenterpb.ListFindingsRequest{
		Parent: sourceName,
	}

	var results []*securitycenterpb.ListFindingsResponse_ListFindingsResult
	it := c.sccClient.ListFindings(ctx, req)
	for {
		r, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list findings for %s: %w", sourceName, err)
		}
		results = append(results, r)
	}
	return results, nil
}
