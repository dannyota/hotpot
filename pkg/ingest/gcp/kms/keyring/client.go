package keyring

import (
	"context"
	"fmt"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the GCP Cloud KMS API for key rings.
type Client struct {
	kmsClient *kms.KeyManagementClient
}

// NewClient creates a new GCP Cloud KMS client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	kmsClient, err := kms.NewKeyManagementClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create KMS client: %w", err)
	}

	return &Client{kmsClient: kmsClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.kmsClient != nil {
		return c.kmsClient.Close()
	}
	return nil
}

// ListKeyRings lists all key rings in a project for the given locations.
func (c *Client) ListKeyRings(ctx context.Context, projectID string, locations []string) ([]*kmspb.KeyRing, error) {
	var keyRings []*kmspb.KeyRing

	for _, location := range locations {
		parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)
		req := &kmspb.ListKeyRingsRequest{Parent: parent}

		it := c.kmsClient.ListKeyRings(ctx, req)
		for {
			kr, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				// Skip locations that return errors
				break
			}
			keyRings = append(keyRings, kr)
		}
	}

	return keyRings, nil
}
