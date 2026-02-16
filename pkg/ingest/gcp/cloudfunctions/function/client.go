package function

import (
	"context"
	"fmt"

	functions "cloud.google.com/go/functions/apiv2"
	"cloud.google.com/go/functions/apiv2/functionspb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps the GCP Cloud Functions API.
type Client struct {
	fnClient *functions.FunctionClient
}

// NewClient creates a new Cloud Functions client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	fnClient, err := functions.NewFunctionClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Functions client: %w", err)
	}
	return &Client{fnClient: fnClient}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.fnClient != nil {
		return c.fnClient.Close()
	}
	return nil
}

// ListFunctions lists all Cloud Functions in a project across all locations.
func (c *Client) ListFunctions(ctx context.Context, projectID string) ([]*functionspb.Function, error) {
	var funcs []*functionspb.Function

	parent := "projects/" + projectID + "/locations/-"
	req := &functionspb.ListFunctionsRequest{Parent: parent}

	it := c.fnClient.ListFunctions(ctx, req)
	for {
		f, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list functions in project %s: %w", projectID, err)
		}
		funcs = append(funcs, f)
	}

	return funcs, nil
}
