package dnspolicy

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/api/option"

	dnsv1 "google.golang.org/api/dns/v1"
)

// Client wraps GCP Cloud DNS API for policies.
type Client struct {
	service *dnsv1.Service
}

// NewClient creates a new GCP Cloud DNS policy client.
func NewClient(ctx context.Context, httpClient *http.Client, opts ...option.ClientOption) (*Client, error) {
	allOpts := append([]option.ClientOption{option.WithHTTPClient(httpClient)}, opts...)
	service, err := dnsv1.NewService(ctx, allOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create DNS service: %w", err)
	}

	return &Client{
		service: service,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	// REST clients don't need explicit close
	return nil
}

// ListPolicies lists all DNS policies in a project.
func (c *Client) ListPolicies(ctx context.Context, projectID string) ([]*dnsv1.Policy, error) {
	var policies []*dnsv1.Policy

	call := c.service.Policies.List(projectID)
	err := call.Pages(ctx, func(resp *dnsv1.PoliciesListResponse) error {
		policies = append(policies, resp.Policies...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list policies in project %s: %w", projectID, err)
	}

	return policies, nil
}
