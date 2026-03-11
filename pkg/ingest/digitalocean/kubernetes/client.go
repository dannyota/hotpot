package kubernetes

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
)

// Client wraps the DigitalOcean Kubernetes API with pagination.
type Client struct {
	godoClient *godo.Client
}

// NewClient creates a new DigitalOcean Kubernetes client.
func NewClient(godoClient *godo.Client) *Client {
	return &Client{godoClient: godoClient}
}

// ListAllClusters fetches all Kubernetes clusters using page-based pagination.
func (c *Client) ListAllClusters(ctx context.Context) ([]*godo.KubernetesCluster, error) {
	var all []*godo.KubernetesCluster
	opt := &godo.ListOptions{Page: 1, PerPage: 200}
	for {
		clusters, resp, err := c.godoClient.Kubernetes.List(ctx, opt)
		if err != nil {
			return nil, fmt.Errorf("list kubernetes clusters (page %d): %w", opt.Page, err)
		}
		all = append(all, clusters...)
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}
		opt.Page++
	}
	return all, nil
}

// ListAllNodePools fetches all node pools for a Kubernetes cluster using page-based pagination.
func (c *Client) ListAllNodePools(ctx context.Context, clusterID string) ([]*godo.KubernetesNodePool, error) {
	var all []*godo.KubernetesNodePool
	opt := &godo.ListOptions{Page: 1, PerPage: 200}
	for {
		pools, resp, err := c.godoClient.Kubernetes.ListNodePools(ctx, clusterID, opt)
		if err != nil {
			return nil, fmt.Errorf("list node pools for cluster %s (page %d): %w", clusterID, opt.Page, err)
		}
		all = append(all, pools...)
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}
		opt.Page++
	}
	return all, nil
}
