package logbucket

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

// IngestLoggingBucketsParams contains parameters for the ingest activity.
type IngestLoggingBucketsParams struct {
	ProjectID string
}

// IngestLoggingBucketsResult contains the result of the ingest activity.
type IngestLoggingBucketsResult struct {
	ProjectID      string
	BucketCount    int
	DurationMillis int64
}

// IngestLoggingBucketsActivity is the activity function reference for workflow registration.
var IngestLoggingBucketsActivity = (*Activities).IngestLoggingBuckets

// IngestLoggingBuckets is a Temporal activity that ingests GCP Cloud Logging buckets.
func (a *Activities) IngestLoggingBuckets(ctx context.Context, params IngestLoggingBucketsParams) (*IngestLoggingBucketsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting Logging bucket ingestion", "projectID", params.ProjectID)

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{ProjectID: params.ProjectID})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest buckets: %w", err)
	}

	if err := service.DeleteStaleBuckets(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale buckets", "error", err)
	}

	logger.Info("Completed Logging bucket ingestion",
		"projectID", params.ProjectID,
		"bucketCount", result.BucketCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestLoggingBucketsResult{
		ProjectID:      result.ProjectID,
		BucketCount:    result.BucketCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
