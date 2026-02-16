package database

import (
	"context"
	"fmt"
	"net/http"

	"github.com/digitalocean/godo"
	"go.temporal.io/sdk/activity"
	"golang.org/x/oauth2"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *ent.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		limiter:       limiter,
	}
}

func (a *Activities) createClient() *Client {
	rateLimitedTransport := ratelimit.NewRateLimitedTransport(a.limiter, nil)
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: a.configService.DOAPIToken()})
	oauthTransport := &oauth2.Transport{Source: tokenSource, Base: rateLimitedTransport}
	httpClient := &http.Client{Transport: oauthTransport}
	godoClient := godo.NewClient(httpClient)
	return NewClient(godoClient)
}

// IngestDODatabasesResult contains the result of the databases ingest activity.
type IngestDODatabasesResult struct {
	ClusterCount   int
	ClusterIDs     []string
	EngineMap      map[string]string
	DurationMillis int64
}

// IngestDODatabasesActivity is the activity function reference for workflow registration.
var IngestDODatabasesActivity = (*Activities).IngestDODatabases

// IngestDODatabases is a Temporal activity that ingests DigitalOcean Database clusters.
func (a *Activities) IngestDODatabases(ctx context.Context) (*IngestDODatabasesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting DigitalOcean Database ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.IngestDatabases(ctx, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("ingest databases: %w", err)
	}

	if err := service.DeleteStaleDatabases(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale databases", "error", err)
	}

	logger.Info("Completed DigitalOcean Database ingestion",
		"clusterCount", result.ClusterCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestDODatabasesResult{
		ClusterCount:   result.ClusterCount,
		ClusterIDs:     result.ClusterIDs,
		EngineMap:      result.EngineMap,
		DurationMillis: result.DurationMillis,
	}, nil
}

// IngestDODatabaseChildrenInput contains the input for the children ingest activity.
type IngestDODatabaseChildrenInput struct {
	ClusterIDs []string
	EngineMap  map[string]string
}

// IngestDODatabaseChildrenResult contains the result of the children ingest activity.
type IngestDODatabaseChildrenResult struct {
	FirewallRuleCount int
	UserCount         int
	ReplicaCount      int
	BackupCount       int
	ConfigCount       int
	PoolCount         int
	DurationMillis    int64
}

// IngestDODatabaseChildrenActivity is the activity function reference for workflow registration.
var IngestDODatabaseChildrenActivity = (*Activities).IngestDODatabaseChildren

// IngestDODatabaseChildren is a Temporal activity that ingests all child resources for Database clusters.
func (a *Activities) IngestDODatabaseChildren(ctx context.Context, input IngestDODatabaseChildrenInput) (*IngestDODatabaseChildrenResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting DigitalOcean Database children ingestion", "clusterCount", len(input.ClusterIDs))

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.IngestChildren(ctx, input.ClusterIDs, input.EngineMap, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("ingest database children: %w", err)
	}

	if err := service.DeleteStaleFirewallRules(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale firewall rules", "error", err)
	}
	if err := service.DeleteStaleUsers(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale users", "error", err)
	}
	if err := service.DeleteStaleReplicas(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale replicas", "error", err)
	}
	if err := service.DeleteStaleBackups(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale backups", "error", err)
	}
	if err := service.DeleteStaleConfigs(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale configs", "error", err)
	}
	if err := service.DeleteStalePools(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale pools", "error", err)
	}

	logger.Info("Completed DigitalOcean Database children ingestion",
		"firewallRuleCount", result.FirewallRuleCount,
		"userCount", result.UserCount,
		"replicaCount", result.ReplicaCount,
		"backupCount", result.BackupCount,
		"configCount", result.ConfigCount,
		"poolCount", result.PoolCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestDODatabaseChildrenResult{
		FirewallRuleCount: result.FirewallRuleCount,
		UserCount:         result.UserCount,
		ReplicaCount:      result.ReplicaCount,
		BackupCount:       result.BackupCount,
		ConfigCount:       result.ConfigCount,
		PoolCount:         result.PoolCount,
		DurationMillis:    result.DurationMillis,
	}, nil
}
