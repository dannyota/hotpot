package managedzone

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/api/option"

	dnsv1 "google.golang.org/api/dns/v1"
)

// Client wraps GCP Cloud DNS API for managed zones.
type Client struct {
	service *dnsv1.Service
}

// NewClient creates a new GCP Cloud DNS managed zone client.
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

// ListManagedZones lists all managed zones in a project.
func (c *Client) ListManagedZones(ctx context.Context, projectID string) ([]*dnsv1.ManagedZone, error) {
	var zones []*dnsv1.ManagedZone

	call := c.service.ManagedZones.List(projectID)
	err := call.Pages(ctx, func(resp *dnsv1.ManagedZonesListResponse) error {
		zones = append(zones, resp.ManagedZones...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list managed zones in project %s: %w", projectID, err)
	}

	return zones, nil
}
