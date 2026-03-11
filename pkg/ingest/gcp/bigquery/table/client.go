package table

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the GCP BigQuery API for tables.
type Client struct {
	bqClient *bigquery.Client
}

// NewClient creates a new BigQuery table client.
func NewClient(ctx context.Context, projectID string, opts ...option.ClientOption) (*Client, error) {
	bqClient, err := bigquery.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create bigquery client: %w", err)
	}
	return &Client{bqClient: bqClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.bqClient != nil {
		return c.bqClient.Close()
	}
	return nil
}

// TableRaw holds raw API data for a BigQuery table.
type TableRaw struct {
	DatasetID string
	TableID   string
	Metadata  *bigquery.TableMetadata
}

// ListTables lists all tables across the given datasets.
func (c *Client) ListTables(ctx context.Context, datasetIDs []string) ([]TableRaw, error) {
	var tables []TableRaw
	for _, dsID := range datasetIDs {
		dsTables, err := c.listTablesForDataset(ctx, dsID)
		if err != nil {
			// Skip individual dataset failures
			continue
		}
		tables = append(tables, dsTables...)
	}
	return tables, nil
}

// listTablesForDataset lists all tables in a single dataset.
func (c *Client) listTablesForDataset(ctx context.Context, datasetID string) ([]TableRaw, error) {
	ds := c.bqClient.Dataset(datasetID)
	it := ds.Tables(ctx)

	var tables []TableRaw
	for {
		tbl, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list tables in dataset %s: %w", datasetID, err)
		}

		meta, err := tbl.Metadata(ctx)
		if err != nil {
			// Skip individual table failures
			continue
		}

		tables = append(tables, TableRaw{
			DatasetID: datasetID,
			TableID:   tbl.TableID,
			Metadata:  meta,
		})
	}
	return tables, nil
}
