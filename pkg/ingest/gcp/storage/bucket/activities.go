package bucket

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/gcpauth"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/base/temporalerr"
	entstorage "github.com/dannyota/hotpot/pkg/storage/ent/gcp/storage"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entstorage.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entstorage.Client, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		limiter:       limiter,
	}
}

// createClient creates a rate-limited GCP client with credentials.
func (a *Activities) createClient(ctx context.Context) (*Client, error) {
	httpClient, err := gcpauth.NewHTTPClient(ctx, a.configService.GCPCredentialsJSON(), a.limiter)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, option.WithHTTPClient(httpClient))
}

// IngestStorageBucketsParams contains parameters for the ingest activity.
type IngestStorageBucketsParams struct {
	ProjectID string
}

// IngestStorageBucketsResult contains the result of the ingest activity.
type IngestStorageBucketsResult struct {
	ProjectID      string
	BucketCount    int
	DurationMillis int64
}

// IngestStorageBucketsActivity is the activity function reference for workflow registration.
var IngestStorageBucketsActivity = (*Activities).IngestStorageBuckets

// IngestStorageBuckets is a Temporal activity that ingests GCP Storage buckets.
func (a *Activities) IngestStorageBuckets(ctx context.Context, params IngestStorageBucketsParams) (*IngestStorageBucketsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Storage bucket ingestion",
		"projectID", params.ProjectID,
	)

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest buckets: %w", err))
	}

	if err := service.DeleteStaleBuckets(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale buckets", "error", err)
	}

	logger.Info("Completed GCP Storage bucket ingestion",
		"projectID", params.ProjectID,
		"bucketCount", result.BucketCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestStorageBucketsResult{
		ProjectID:      result.ProjectID,
		BucketCount:    result.BucketCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
