package image

import (
	"context"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client wraps GCP Compute Engine API for images.
type Client struct {
	imagesClient *compute.ImagesClient
}

// NewClient creates a new GCP Compute image client.
func NewClient(ctx context.Context, opts ...option.ClientOption) (*Client, error) {
	imagesClient, err := compute.NewImagesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create images client: %w", err)
	}

	return &Client{
		imagesClient: imagesClient,
	}, nil
}

// Close closes the client connections.
func (c *Client) Close() error {
	if c.imagesClient != nil {
		return c.imagesClient.Close()
	}
	return nil
}

// ListImages lists all images in a project (global resource).
func (c *Client) ListImages(ctx context.Context, projectID string) ([]*computepb.Image, error) {
	req := &computepb.ListImagesRequest{
		Project: projectID,
	}

	var images []*computepb.Image
	it := c.imagesClient.List(ctx, req)

	for {
		img, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list images in project %s: %w", projectID, err)
		}

		images = append(images, img)
	}

	return images, nil
}
