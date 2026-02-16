package kubernetes

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

// IngestDOKubernetesClustersResult contains the result of the clusters ingest activity.
type IngestDOKubernetesClustersResult struct {
	ClusterCount   int
	ClusterIDs     []string
	DurationMillis int64
}

// IngestDOKubernetesClustersActivity is the activity function reference for workflow registration.
var IngestDOKubernetesClustersActivity = (*Activities).IngestDOKubernetesClusters

// IngestDOKubernetesClusters is a Temporal activity that ingests DigitalOcean Kubernetes clusters.
func (a *Activities) IngestDOKubernetesClusters(ctx context.Context) (*IngestDOKubernetesClustersResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting DigitalOcean Kubernetes cluster ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.IngestClusters(ctx, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("ingest kubernetes clusters: %w", err)
	}

	if err := service.DeleteStaleClusters(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale kubernetes clusters", "error", err)
	}

	logger.Info("Completed DigitalOcean Kubernetes cluster ingestion",
		"clusterCount", result.ClusterCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestDOKubernetesClustersResult{
		ClusterCount:   result.ClusterCount,
		ClusterIDs:     result.ClusterIDs,
		DurationMillis: result.DurationMillis,
	}, nil
}

// IngestDOKubernetesNodePoolsInput contains the input for the node pools ingest activity.
type IngestDOKubernetesNodePoolsInput struct {
	ClusterIDs []string
}

// IngestDOKubernetesNodePoolsResult contains the result of the node pools ingest activity.
type IngestDOKubernetesNodePoolsResult struct {
	NodePoolCount  int
	DurationMillis int64
}

// IngestDOKubernetesNodePoolsActivity is the activity function reference for workflow registration.
var IngestDOKubernetesNodePoolsActivity = (*Activities).IngestDOKubernetesNodePools

// IngestDOKubernetesNodePools is a Temporal activity that ingests node pools for Kubernetes clusters.
func (a *Activities) IngestDOKubernetesNodePools(ctx context.Context, input IngestDOKubernetesNodePoolsInput) (*IngestDOKubernetesNodePoolsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting DigitalOcean Kubernetes node pool ingestion", "clusterCount", len(input.ClusterIDs))

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.IngestNodePools(ctx, input.ClusterIDs, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("ingest kubernetes node pools: %w", err)
	}

	if err := service.DeleteStaleNodePools(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale kubernetes node pools", "error", err)
	}

	logger.Info("Completed DigitalOcean Kubernetes node pool ingestion",
		"nodePoolCount", result.NodePoolCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestDOKubernetesNodePoolsResult{
		NodePoolCount:  result.NodePoolCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
