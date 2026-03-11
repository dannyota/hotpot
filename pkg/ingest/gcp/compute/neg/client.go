package neg

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Client struct {
	negClient *compute.NetworkEndpointGroupsClient
}

func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	c, err := compute.NewNetworkEndpointGroupsRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create NEG client: %w", err)
	}
	return &Client{negClient: c}, nil
}

func (c *Client) Close() error {
	if c.negClient != nil {
		return c.negClient.Close()
	}
	return nil
}

func (c *Client) ListNegs(ctx context.Context, projectID string) ([]*computepb.NetworkEndpointGroup, error) {
	req := &computepb.AggregatedListNetworkEndpointGroupsRequest{
		Project: projectID,
	}

	var negs []*computepb.NetworkEndpointGroup
	it := c.negClient.AggregatedList(ctx, req)

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list NEGs in project %s: %w", projectID, err)
		}

		negs = append(negs, pair.Value.NetworkEndpointGroups...)
	}

	return negs, nil
}
