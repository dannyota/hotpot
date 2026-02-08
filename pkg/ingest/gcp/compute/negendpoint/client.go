package negendpoint

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzegcpcomputeneg"
)

// EndpointWithNeg pairs an endpoint with its parent NEG info.
type EndpointWithNeg struct {
	Endpoint      *computepb.NetworkEndpointWithHealthStatus
	NegResourceID string
	NegName       string
	Zone          string
}

type Client struct {
	negClient *compute.NetworkEndpointGroupsClient
	entClient *ent.Client
}

func NewClient(ctx context.Context, entClient *ent.Client, opts ...option.ClientOption) (*Client, error) {
	c, err := compute.NewNetworkEndpointGroupsRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create NEG client: %w", err)
	}
	return &Client{negClient: c, entClient: entClient}, nil
}

func (c *Client) Close() error {
	if c.negClient != nil {
		return c.negClient.Close()
	}
	return nil
}

// ListNegEndpoints queries NEGs from the database, then lists endpoints for each.
func (c *Client) ListNegEndpoints(ctx context.Context, projectID string) ([]EndpointWithNeg, error) {
	// Query all NEGs for this project from the database
	negs, err := c.entClient.BronzeGCPComputeNeg.Query().
		Where(bronzegcpcomputeneg.ProjectID(projectID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query NEGs from database: %w", err)
	}

	var endpoints []EndpointWithNeg
	for _, neg := range negs {
		// Only zonal NEGs with certain types support ListNetworkEndpoints
		if neg.Zone == "" {
			continue
		}

		// Extract zone name from zone URL (e.g., ".../zones/us-central1-a" -> "us-central1-a")
		zoneName := extractZoneName(neg.Zone)

		req := &computepb.ListNetworkEndpointsNetworkEndpointGroupsRequest{
			Project:              projectID,
			Zone:                 zoneName,
			NetworkEndpointGroup: neg.Name,
		}

		it := c.negClient.ListNetworkEndpoints(ctx, req)
		for {
			ep, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				// Some NEG types don't support listing endpoints (e.g., INTERNET_IP_PORT, SERVERLESS)
				// Skip gracefully
				break
			}
			endpoints = append(endpoints, EndpointWithNeg{
				Endpoint:      ep,
				NegResourceID: neg.ID,
				NegName:       neg.Name,
				Zone:          zoneName,
			})
		}
	}

	return endpoints, nil
}

// extractZoneName extracts the zone name from a full zone URL.
func extractZoneName(zone string) string {
	// Zone could be a full URL like "https://www.googleapis.com/compute/v1/projects/PROJECT/zones/ZONE"
	// or just a zone name like "us-central1-a"
	for i := len(zone) - 1; i >= 0; i-- {
		if zone[i] == '/' {
			return zone[i+1:]
		}
	}
	return zone
}
