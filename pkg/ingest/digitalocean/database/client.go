package database

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/digitalocean/godo"
)

// Client wraps the DigitalOcean Databases API with pagination.
type Client struct {
	godoClient *godo.Client
}

// NewClient creates a new DigitalOcean Database client.
func NewClient(godoClient *godo.Client) *Client {
	return &Client{godoClient: godoClient}
}

// ListAllDatabases fetches all database clusters using page-based pagination.
func (c *Client) ListAllDatabases(ctx context.Context) ([]godo.Database, error) {
	var all []godo.Database
	opt := &godo.ListOptions{Page: 1, PerPage: 200}
	for {
		databases, resp, err := c.godoClient.Databases.List(ctx, opt)
		if err != nil {
			return nil, fmt.Errorf("list databases (page %d): %w", opt.Page, err)
		}
		all = append(all, databases...)
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}
		opt.Page++
	}
	return all, nil
}

// GetFirewallRules fetches firewall rules for a database cluster.
func (c *Client) GetFirewallRules(ctx context.Context, clusterID string) ([]godo.DatabaseFirewallRule, error) {
	rules, _, err := c.godoClient.Databases.GetFirewallRules(ctx, clusterID)
	if err != nil {
		return nil, fmt.Errorf("get firewall rules for cluster %s: %w", clusterID, err)
	}
	return rules, nil
}

// ListAllUsers fetches all users for a database cluster using page-based pagination.
func (c *Client) ListAllUsers(ctx context.Context, clusterID string) ([]godo.DatabaseUser, error) {
	var all []godo.DatabaseUser
	opt := &godo.ListOptions{Page: 1, PerPage: 200}
	for {
		users, resp, err := c.godoClient.Databases.ListUsers(ctx, clusterID, opt)
		if err != nil {
			return nil, fmt.Errorf("list users for cluster %s (page %d): %w", clusterID, opt.Page, err)
		}
		all = append(all, users...)
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}
		opt.Page++
	}
	return all, nil
}

// ListAllReplicas fetches all replicas for a database cluster using page-based pagination.
func (c *Client) ListAllReplicas(ctx context.Context, clusterID string) ([]godo.DatabaseReplica, error) {
	var all []godo.DatabaseReplica
	opt := &godo.ListOptions{Page: 1, PerPage: 200}
	for {
		replicas, resp, err := c.godoClient.Databases.ListReplicas(ctx, clusterID, opt)
		if err != nil {
			return nil, fmt.Errorf("list replicas for cluster %s (page %d): %w", clusterID, opt.Page, err)
		}
		all = append(all, replicas...)
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}
		opt.Page++
	}
	return all, nil
}

// ListAllBackups fetches all backups for a database cluster using page-based pagination.
func (c *Client) ListAllBackups(ctx context.Context, clusterID string) ([]godo.DatabaseBackup, error) {
	var all []godo.DatabaseBackup
	opt := &godo.ListOptions{Page: 1, PerPage: 200}
	for {
		backups, resp, err := c.godoClient.Databases.ListBackups(ctx, clusterID, opt)
		if err != nil {
			return nil, fmt.Errorf("list backups for cluster %s (page %d): %w", clusterID, opt.Page, err)
		}
		all = append(all, backups...)
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}
		opt.Page++
	}
	return all, nil
}

// GetConfig fetches the engine-specific config for a database cluster as a JSON blob.
func (c *Client) GetConfig(ctx context.Context, clusterID, engineSlug string) (json.RawMessage, error) {
	var cfg any
	var err error

	switch engineSlug {
	case "pg":
		cfg, _, err = c.godoClient.Databases.GetPostgreSQLConfig(ctx, clusterID)
	case "mysql":
		cfg, _, err = c.godoClient.Databases.GetMySQLConfig(ctx, clusterID)
	case "redis":
		cfg, _, err = c.godoClient.Databases.GetRedisConfig(ctx, clusterID)
	case "valkey":
		cfg, _, err = c.godoClient.Databases.GetValkeyConfig(ctx, clusterID)
	case "mongodb":
		cfg, _, err = c.godoClient.Databases.GetMongoDBConfig(ctx, clusterID)
	case "kafka":
		cfg, _, err = c.godoClient.Databases.GetKafkaConfig(ctx, clusterID)
	case "opensearch":
		cfg, _, err = c.godoClient.Databases.GetOpensearchConfig(ctx, clusterID)
	default:
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get %s config for cluster %s: %w", engineSlug, clusterID, err)
	}
	if cfg == nil {
		return nil, nil
	}

	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal %s config for cluster %s: %w", engineSlug, clusterID, err)
	}
	return raw, nil
}

// ListAllPools fetches all connection pools for a database cluster using page-based pagination.
func (c *Client) ListAllPools(ctx context.Context, clusterID string) ([]godo.DatabasePool, error) {
	var all []godo.DatabasePool
	opt := &godo.ListOptions{Page: 1, PerPage: 200}
	for {
		pools, resp, err := c.godoClient.Databases.ListPools(ctx, clusterID, opt)
		if err != nil {
			return nil, fmt.Errorf("list pools for cluster %s (page %d): %w", clusterID, opt.Page, err)
		}
		all = append(all, pools...)
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}
		opt.Page++
	}
	return all, nil
}
