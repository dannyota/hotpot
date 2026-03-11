package folder

import (
	"context"
	"fmt"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Cloud Resource Manager API for folders.
type Client struct {
	foldersClient *resourcemanager.FoldersClient
}

// NewClient creates a new GCP Resource Manager folders client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	foldersClient, err := resourcemanager.NewFoldersClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create folders client: %w", err)
	}

	return &Client{
		foldersClient: foldersClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.foldersClient != nil {
		return c.foldersClient.Close()
	}
	return nil
}

// SearchFolders searches for all folders accessible by the service account.
// Returns folders where the caller has resourcemanager.folders.get permission.
func (c *Client) SearchFolders(ctx context.Context) ([]*resourcemanagerpb.Folder, error) {
	req := &resourcemanagerpb.SearchFoldersRequest{}

	var folders []*resourcemanagerpb.Folder
	it := c.foldersClient.SearchFolders(ctx, req)

	for {
		folder, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to search folders: %w", err)
		}

		folders = append(folders, folder)
	}

	return folders, nil
}
