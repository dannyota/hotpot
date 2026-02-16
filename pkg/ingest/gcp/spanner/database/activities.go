package database

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

// IngestSpannerDatabasesParams contains parameters for the ingest activity.
type IngestSpannerDatabasesParams struct {
	ProjectID     string
	InstanceNames []string
}

// IngestSpannerDatabasesResult contains the result of the ingest activity.
type IngestSpannerDatabasesResult struct {
	ProjectID      string
	DatabaseCount  int
	DurationMillis int64
}

// IngestSpannerDatabasesActivity is the activity function reference for workflow registration.
var IngestSpannerDatabasesActivity = (*Activities).IngestSpannerDatabases

// IngestSpannerDatabases is a Temporal activity that ingests GCP Spanner databases.
func (a *Activities) IngestSpannerDatabases(ctx context.Context, params IngestSpannerDatabasesParams) (*IngestSpannerDatabasesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting Spanner database ingestion",
		"projectID", params.ProjectID,
		"instanceCount", len(params.InstanceNames),
	)

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID:     params.ProjectID,
		InstanceNames: params.InstanceNames,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest Spanner databases: %w", err)
	}

	if err := service.DeleteStaleDatabases(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale Spanner databases", "error", err)
	}

	logger.Info("Completed Spanner database ingestion",
		"projectID", params.ProjectID,
		"databaseCount", result.DatabaseCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestSpannerDatabasesResult{
		ProjectID:      result.ProjectID,
		DatabaseCount:  result.DatabaseCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
