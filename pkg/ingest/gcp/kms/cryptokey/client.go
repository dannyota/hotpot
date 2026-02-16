package cryptokey

import (
	"context"
	"fmt"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the GCP Cloud KMS API for crypto keys.
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

// ListCryptoKeys lists all crypto keys in a key ring.
func (c *Client) ListCryptoKeys(ctx context.Context, keyRingName string) ([]*kmspb.CryptoKey, error) {
	var keys []*kmspb.CryptoKey

	req := &kmspb.ListCryptoKeysRequest{Parent: keyRingName}
	it := c.kmsClient.ListCryptoKeys(ctx, req)

	for {
		key, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list crypto keys in key ring %s: %w", keyRingName, err)
		}
		keys = append(keys, key)
	}

	return keys, nil
}
