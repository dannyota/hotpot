package cluster

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

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

// createClient creates a rate-limited GCP client with credentials.
func (a *Activities) createClient(ctx context.Context) (*Client, error) {
	var opts []option.ClientOption
	if credJSON := a.configService.GCPCredentialsJSON(); len(credJSON) > 0 {
		opts = append(opts, option.WithAuthCredentialsJSON(option.ServiceAccount, credJSON))
	}
	opts = append(opts, option.WithGRPCDialOption(
		grpc.WithUnaryInterceptor(ratelimit.UnaryInterceptor(a.limiter)),
	))
	return NewClient(ctx, opts...)
}

// IngestAlloyDBClustersParams contains parameters for the ingest activity.
type IngestAlloyDBClustersParams struct {
	ProjectID string
}

// IngestAlloyDBClustersResult contains the result of the ingest activity.
type IngestAlloyDBClustersResult struct {
	ProjectID      string
	ClusterCount   int
	DurationMillis int64
}

// IngestAlloyDBClustersActivity is the activity function reference for workflow registration.
var IngestAlloyDBClustersActivity = (*Activities).IngestAlloyDBClusters

// IngestAlloyDBClusters is a Temporal activity that ingests AlloyDB clusters.
func (a *Activities) IngestAlloyDBClusters(ctx context.Context, params IngestAlloyDBClustersParams) (*IngestAlloyDBClustersResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting AlloyDB cluster ingestion",
		"projectID", params.ProjectID,
	)

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest AlloyDB clusters: %w", err)
	}

	if err := service.DeleteStaleClusters(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale AlloyDB clusters", "error", err)
	}

	logger.Info("Completed AlloyDB cluster ingestion",
		"projectID", params.ProjectID,
		"clusterCount", result.ClusterCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestAlloyDBClustersResult{
		ProjectID:      result.ProjectID,
		ClusterCount:   result.ClusterCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
