package note

import (
	"context"
	"fmt"

	containeranalysis "cloud.google.com/go/containeranalysis/apiv1"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	grafeaspb "google.golang.org/genproto/googleapis/grafeas/v1"
)

// Client wraps the GCP Container Analysis / Grafeas API for notes.
type Client struct {
	caClient *containeranalysis.Client
}

// NewClient creates a new Container Analysis client for notes.
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

// ListNotes lists all Grafeas notes in a project.
func (c *Client) ListNotes(ctx context.Context, projectID string) ([]*grafeaspb.Note, error) {
	grafeasClient := c.caClient.GetGrafeasClient()

	req := &grafeaspb.ListNotesRequest{
		Parent: "projects/" + projectID,
	}

	var notes []*grafeaspb.Note
	it := grafeasClient.ListNotes(ctx, req)
	for {
		n, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list notes for project %s: %w", projectID, err)
		}
		notes = append(notes, n)
	}
	return notes, nil
}
