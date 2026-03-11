package alertpolicy

import (
	"context"
	"fmt"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the GCP Monitoring API for alert policies.
type Client struct {
	alertPolicyClient *monitoring.AlertPolicyClient
}

// NewClient creates a new Monitoring alert policy client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	c, err := monitoring.NewAlertPolicyClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create alert policy client: %w", err)
	}
	return &Client{alertPolicyClient: c}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.alertPolicyClient != nil {
		return c.alertPolicyClient.Close()
	}
	return nil
}

// ListAlertPolicies lists all alert policies in a project.
func (c *Client) ListAlertPolicies(ctx context.Context, projectID string) ([]*monitoringpb.AlertPolicy, error) {
	req := &monitoringpb.ListAlertPoliciesRequest{
		Name: "projects/" + projectID,
	}

	var policies []*monitoringpb.AlertPolicy
	it := c.alertPolicyClient.ListAlertPolicies(ctx, req)
	for {
		p, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list alert policies for project %s: %w", projectID, err)
		}
		policies = append(policies, p)
	}
	return policies, nil
}
