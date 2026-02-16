package revision

import (
	"context"
	"fmt"

	run "cloud.google.com/go/run/apiv2"
	"cloud.google.com/go/run/apiv2/runpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcprunservice"
)

// Client wraps the GCP Cloud Run API for revisions.
type Client struct {
	revisionsClient *run.RevisionsClient
	entClient       *ent.Client
}

// NewClient creates a new Cloud Run revision client.
func NewClient(ctx context.Context, entClient *ent.Client, opts ...option.ClientOption) (*Client, error) {
	revisionsClient, err := run.NewRevisionsClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Run revisions client: %w", err)
	}
	return &Client{revisionsClient: revisionsClient, entClient: entClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.revisionsClient != nil {
		return c.revisionsClient.Close()
	}
	return nil
}

// RevisionRaw holds raw API data for a Cloud Run revision.
type RevisionRaw struct {
	ServiceName string
	Revision    *runpb.Revision
}

// ListRevisions queries services from the database for the given project,
// then fetches revisions for each service.
func (c *Client) ListRevisions(ctx context.Context, projectID string) ([]RevisionRaw, error) {
	// Query services from database for this project
	services, err := c.entClient.BronzeGCPRunService.Query().
		Where(bronzegcprunservice.ProjectID(projectID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query Cloud Run services from database: %w", err)
	}

	var revisions []RevisionRaw
	for _, svc := range services {
		svcRevisions, err := c.listRevisionsForService(ctx, svc.ID)
		if err != nil {
			// Skip individual service failures
			continue
		}
		for _, rev := range svcRevisions {
			revisions = append(revisions, RevisionRaw{
				ServiceName: svc.ID,
				Revision:    rev,
			})
		}
	}
	return revisions, nil
}

// listRevisionsForService fetches all revisions for a single Cloud Run service.
func (c *Client) listRevisionsForService(ctx context.Context, serviceName string) ([]*runpb.Revision, error) {
	req := &runpb.ListRevisionsRequest{
		Parent: serviceName,
	}

	var revisions []*runpb.Revision
	it := c.revisionsClient.ListRevisions(ctx, req)
	for {
		rev, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list revisions for service %s: %w", serviceName, err)
		}
		revisions = append(revisions, rev)
	}
	return revisions, nil
}
