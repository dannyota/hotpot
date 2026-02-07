package serviceaccount

import (
	"context"
	"fmt"

	admin "cloud.google.com/go/iam/admin/apiv1"
	"cloud.google.com/go/iam/admin/apiv1/adminpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

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

func (c *Client) ListServiceAccounts(ctx context.Context, projectID string) ([]*adminpb.ServiceAccount, error) {
	req := &adminpb.ListServiceAccountsRequest{
		Name: "projects/" + projectID,
	}

	var accounts []*adminpb.ServiceAccount
	it := c.iamClient.ListServiceAccounts(ctx, req)
	for {
		sa, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list service accounts in project %s: %w", projectID, err)
		}
		accounts = append(accounts, sa)
	}
	return accounts, nil
}
