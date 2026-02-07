package cluster

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"golang.org/x/time/rate"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	db            *gorm.DB
	limiter       *rate.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, db *gorm.DB, limiter *rate.Limiter) *Activities {
	return &Activities{
		configService: configService,
		db:            db,
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

// IngestContainerClustersParams contains parameters for the ingest activity.
type IngestContainerClustersParams struct {
	ProjectID string
}

// IngestContainerClustersResult contains the result of the ingest activity.
type IngestContainerClustersResult struct {
	ProjectID      string
	ClusterCount   int
	DurationMillis int64
}

// IngestContainerClustersActivity is the activity function reference for workflow registration.
var IngestContainerClustersActivity = (*Activities).IngestContainerClusters

// IngestContainerClusters is a Temporal activity that ingests GKE clusters.
func (a *Activities) IngestContainerClusters(ctx context.Context, params IngestContainerClustersParams) (*IngestContainerClustersResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GKE cluster ingestion",
		"projectID", params.ProjectID,
	)

	// Create client for this activity
	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	// Create service
	service := NewService(client, a.db)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest clusters: %w", err)
	}

	// Delete stale clusters
	if err := service.DeleteStaleClusters(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale clusters", "error", err)
	}

	logger.Info("Completed GKE cluster ingestion",
		"projectID", params.ProjectID,
		"clusterCount", result.ClusterCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestContainerClustersResult{
		ProjectID:      result.ProjectID,
		ClusterCount:   result.ClusterCount,
		DurationMillis: result.DurationMillis,
	}, nil
}

