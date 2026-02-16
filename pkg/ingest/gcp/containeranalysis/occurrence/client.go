package occurrence

import (
	"context"
	"fmt"

	containeranalysis "cloud.google.com/go/containeranalysis/apiv1"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	grafeaspb "google.golang.org/genproto/googleapis/grafeas/v1"
)

// Client wraps the GCP Container Analysis / Grafeas API for occurrences.
type Client struct {
	caClient *containeranalysis.Client
}

// NewClient creates a new Container Analysis client for occurrences.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	caClient, err := containeranalysis.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create container analysis client: %w", err)
	}
	return &Client{caClient: caClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.caClient != nil {
		return c.caClient.Close()
	}
	return nil
}

// ListOccurrences lists all Grafeas occurrences in a project.
func (c *Client) ListOccurrences(ctx context.Context, projectID string) ([]*grafeaspb.Occurrence, error) {
	grafeasClient := c.caClient.GetGrafeasClient()

	req := &grafeaspb.ListOccurrencesRequest{
		Parent: "projects/" + projectID,
	}

	var occurrences []*grafeaspb.Occurrence
	it := grafeasClient.ListOccurrences(ctx, req)
	for {
		o, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list occurrences for project %s: %w", projectID, err)
		}
		occurrences = append(occurrences, o)
	}
	return occurrences, nil
}
