package database

import (
	"context"
	"fmt"

	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the GCP Spanner Database Admin API.
type Client struct {
	databaseAdmin *database.DatabaseAdminClient
}

// NewClient creates a new Spanner database admin client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	databaseAdmin, err := database.NewDatabaseAdminClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create spanner database admin client: %w", err)
	}
	return &Client{databaseAdmin: databaseAdmin}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.databaseAdmin != nil {
		return c.databaseAdmin.Close()
	}
	return nil
}

// ListDatabases lists all Spanner databases for a given instance.
func (c *Client) ListDatabases(ctx context.Context, instanceName string) ([]*databasepb.Database, error) {
	req := &databasepb.ListDatabasesRequest{
		Parent: instanceName,
	}

	var databases []*databasepb.Database
	it := c.databaseAdmin.ListDatabases(ctx, req)
	for {
		db, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list spanner databases for instance %s: %w", instanceName, err)
		}
		databases = append(databases, db)
	}
	return databases, nil
}
