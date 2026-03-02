package targetpool

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/gcpauth"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/base/temporalerr"
	entcompute "github.com/dannyota/hotpot/pkg/storage/ent/gcp/compute"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entcompute.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entcompute.Client, limiter ratelimit.Limiter) *Activities {
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

// IngestComputeTargetPoolsParams contains parameters for the ingest activity.
type IngestComputeTargetPoolsParams struct {
	ProjectID string
}

// IngestComputeTargetPoolsResult contains the result of the ingest activity.
type IngestComputeTargetPoolsResult struct {
	ProjectID       string
	TargetPoolCount int
	DurationMillis  int64
}

// IngestComputeTargetPoolsActivity is the activity function reference for workflow registration.
var IngestComputeTargetPoolsActivity = (*Activities).IngestComputeTargetPools

// IngestComputeTargetPools is a Temporal activity that ingests GCP Compute target pools.
func (a *Activities) IngestComputeTargetPools(ctx context.Context, params IngestComputeTargetPoolsParams) (*IngestComputeTargetPoolsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute target pool ingestion",
		"projectID", params.ProjectID,
	)

	// Create client for this activity
	client, err := a.createClient(ctx)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}
	defer client.Close()

	// Create service
	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest target pools: %w", err))
	}

	// Delete stale target pools
	if err := service.DeleteStaleTargetPools(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale target pools", "error", err)
	}

	logger.Info("Completed GCP Compute target pool ingestion",
		"projectID", params.ProjectID,
		"targetPoolCount", result.TargetPoolCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeTargetPoolsResult{
		ProjectID:       result.ProjectID,
		TargetPoolCount: result.TargetPoolCount,
		DurationMillis:  result.DurationMillis,
	}, nil
}
