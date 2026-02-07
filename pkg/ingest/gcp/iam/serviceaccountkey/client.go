package serviceaccountkey

import (
	"context"
	"fmt"

	admin "cloud.google.com/go/iam/admin/apiv1"
	"cloud.google.com/go/iam/admin/apiv1/adminpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// KeyWithAccount pairs a key with its parent service account email.
type KeyWithAccount struct {
	Key                 *adminpb.ServiceAccountKey
	ServiceAccountEmail string
}

type Client struct {
	iamClient *admin.IamClient
}

func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	iamClient, err := admin.NewIamClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create IAM client: %w", err)
	}
	return &Client{iamClient: iamClient}, nil
}

func (c *Client) Close() error {
	if c.iamClient != nil {
		return c.iamClient.Close()
	}
	return nil
}

// ListServiceAccountKeys lists all keys across all service accounts in a project.
func (c *Client) ListServiceAccountKeys(ctx context.Context, projectID string) ([]KeyWithAccount, error) {
	// First, list all service accounts
	saReq := &adminpb.ListServiceAccountsRequest{
		Name: "projects/" + projectID,
	}

	var keys []KeyWithAccount
	it := c.iamClient.ListServiceAccounts(ctx, saReq)
	for {
		sa, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list service accounts in project %s: %w", projectID, err)
		}

		// List keys for this service account
		keyReq := &adminpb.ListServiceAccountKeysRequest{
			Name: sa.GetName(),
		}
		resp, err := c.iamClient.ListServiceAccountKeys(ctx, keyReq)
		if err != nil {
			return nil, fmt.Errorf("list keys for service account %s: %w", sa.GetEmail(), err)
		}

		for _, key := range resp.GetKeys() {
			keys = append(keys, KeyWithAccount{
				Key:                 key,
				ServiceAccountEmail: sa.GetEmail(),
			})
		}
	}

	return keys, nil
}
