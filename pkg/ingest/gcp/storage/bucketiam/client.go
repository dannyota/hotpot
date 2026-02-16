package bucketiam

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/api/option"
	storagev1 "google.golang.org/api/storage/v1"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpstoragebucket"
)

// BucketIamPolicyRaw holds raw API data for a bucket IAM policy.
type BucketIamPolicyRaw struct {
	BucketName string
	Policy     *storagev1.Policy
}

// Client wraps the GCP Storage API for bucket IAM policies.
type Client struct {
	service   *storagev1.Service
	entClient *ent.Client
}

// NewClient creates a new GCP Storage bucket IAM policy client.
func NewClient(ctx context.Context, entClient *ent.Client, httpClient *http.Client, opts ...option.ClientOption) (*Client, error) {
	allOpts := append([]option.ClientOption{option.WithHTTPClient(httpClient)}, opts...)
	service, err := storagev1.NewService(ctx, allOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage service: %w", err)
	}
	return &Client{service: service, entClient: entClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	return nil
}

// ListBucketIamPolicies queries buckets from the database, then fetches IAM policies for each.
func (c *Client) ListBucketIamPolicies(ctx context.Context, projectID string) ([]BucketIamPolicyRaw, error) {
	// Query buckets from database
	buckets, err := c.entClient.BronzeGCPStorageBucket.Query().
		Where(bronzegcpstoragebucket.ProjectID(projectID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query buckets from database: %w", err)
	}

	var policies []BucketIamPolicyRaw
	for _, b := range buckets {
		policy, err := c.service.Buckets.GetIamPolicy(b.Name).Context(ctx).Do()
		if err != nil {
			// Skip individual bucket failures
			continue
		}
		policies = append(policies, BucketIamPolicyRaw{
			BucketName: b.Name,
			Policy:     policy,
		})
	}
	return policies, nil
}
