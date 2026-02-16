package logmetric

import (
	"context"
	"fmt"

	logging "cloud.google.com/go/logging/apiv2"
	"cloud.google.com/go/logging/apiv2/loggingpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Cloud Logging API for log metrics.
type Client struct {
	metricsClient *logging.MetricsClient
}

// NewClient creates a new GCP Cloud Logging metrics client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	metricsClient, err := logging.NewMetricsClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create logging metrics client: %w", err)
	}
	return &Client{metricsClient: metricsClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.metricsClient != nil {
		return c.metricsClient.Close()
	}
	return nil
}

// ListLogMetrics lists all log-based metrics in a project.
func (c *Client) ListLogMetrics(ctx context.Context, projectID string) ([]*loggingpb.LogMetric, error) {
	req := &loggingpb.ListLogMetricsRequest{
		Parent: "projects/" + projectID,
	}

	var metrics []*loggingpb.LogMetric
	it := c.metricsClient.ListLogMetrics(ctx, req)
	for {
		metric, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list log metrics in project %s: %w", projectID, err)
		}
		metrics = append(metrics, metric)
	}
	return metrics, nil
}
