package dataset

import (
	"context"
	"fmt"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the GCP BigQuery API for datasets.
type Client struct {
	bqClient *bigquery.Client
}

// NewClient creates a new BigQuery dataset client.
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

// DatasetRaw holds raw API data for a BigQuery dataset.
type DatasetRaw struct {
	DatasetID string
	Metadata  *bigquery.DatasetMetadata
}

// ListDatasets lists all datasets in the project.
func (c *Client) ListDatasets(ctx context.Context) ([]DatasetRaw, error) {
	var datasets []DatasetRaw
	it := c.bqClient.Datasets(ctx)
	for {
		ds, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list datasets: %w", err)
		}

		meta, err := ds.Metadata(ctx)
		if err != nil {
			// Skip individual dataset failures
			continue
		}

		datasets = append(datasets, DatasetRaw{
			DatasetID: ds.DatasetID,
			Metadata:  meta,
		})
	}
	return datasets, nil
}
