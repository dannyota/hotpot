package instancegroup

import (
	"context"
	"fmt"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for instance groups.
type Client struct {
	instanceGroupsClient *compute.InstanceGroupsClient
}

// NewClient creates a new GCP Compute instance groups client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	igClient, err := compute.NewInstanceGroupsRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance groups client: %w", err)
	}

	return &Client{
		instanceGroupsClient: igClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.instanceGroupsClient != nil {
		return c.instanceGroupsClient.Close()
	}
	return nil
}

// ListInstanceGroups lists all instance groups in a project using aggregated list.
// Returns instance groups from all zones. Skips regional entries (no zone).
func (c *Client) ListInstanceGroups(ctx context.Context, projectID string) ([]*computepb.InstanceGroup, error) {
	req := &computepb.AggregatedListInstanceGroupsRequest{
		Project: projectID,
	}

	var groups []*computepb.InstanceGroup
	it := c.instanceGroupsClient.AggregatedList(ctx, req)

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list instance groups in project %s: %w", projectID, err)
		}

		// Skip entries without zone (regional instance groups use a different API)
		if !strings.HasPrefix(pair.Key, "zones/") {
			continue
		}

		groups = append(groups, pair.Value.InstanceGroups...)
	}

	return groups, nil
}

// ListInstanceGroupMembers lists members of a specific instance group.
func (c *Client) ListInstanceGroupMembers(ctx context.Context, projectID, zone, groupName string) ([]*computepb.InstanceWithNamedPorts, error) {
	req := &computepb.ListInstancesInstanceGroupsRequest{
		Project:       projectID,
		Zone:          zone,
		InstanceGroup: groupName,
		InstanceGroupsListInstancesRequestResource: &computepb.InstanceGroupsListInstancesRequest{},
	}

	var members []*computepb.InstanceWithNamedPorts
	it := c.instanceGroupsClient.ListInstances(ctx, req)

	for {
		member, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list members of instance group %s: %w", groupName, err)
		}

		members = append(members, member)
	}

	return members, nil
}
