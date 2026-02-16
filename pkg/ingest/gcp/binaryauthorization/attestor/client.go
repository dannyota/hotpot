package attestor

import (
	"context"
	"fmt"

	binaryauthorization "cloud.google.com/go/binaryauthorization/apiv1"
	"cloud.google.com/go/binaryauthorization/apiv1/binaryauthorizationpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the GCP Binary Authorization API for attestors.
type Client struct {
	binauthzClient *binaryauthorization.BinauthzManagementClient
}

// NewClient creates a new Binary Authorization attestor client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	c, err := binaryauthorization.NewBinauthzManagementClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create binary authorization client: %w", err)
	}
	return &Client{binauthzClient: c}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.binauthzClient != nil {
		return c.binauthzClient.Close()
	}
	return nil
}

// ListAttestors lists all attestors in a project.
func (c *Client) ListAttestors(ctx context.Context, projectID string) ([]*binaryauthorizationpb.Attestor, error) {
	req := &binaryauthorizationpb.ListAttestorsRequest{
		Parent: "projects/" + projectID,
	}

	var attestors []*binaryauthorizationpb.Attestor
	it := c.binauthzClient.ListAttestors(ctx, req)
	for {
		a, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list attestors in project %s: %w", projectID, err)
		}
		attestors = append(attestors, a)
	}

	return attestors, nil
}
