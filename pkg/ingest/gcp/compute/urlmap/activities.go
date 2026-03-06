package urlmap

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/gcpauth"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	entcompute "danny.vn/hotpot/pkg/storage/ent/gcp/compute"
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

// IngestComputeUrlMapsParams contains parameters for the ingest activity.
type IngestComputeUrlMapsParams struct {
	ProjectID string
}

// IngestComputeUrlMapsResult contains the result of the ingest activity.
type IngestComputeUrlMapsResult struct {
	ProjectID      string
	UrlMapCount    int
	DurationMillis int64
}

// IngestComputeUrlMapsActivity is the activity function reference for workflow registration.
var IngestComputeUrlMapsActivity = (*Activities).IngestComputeUrlMaps

// IngestComputeUrlMaps is a Temporal activity that ingests GCP Compute URL maps.
func (a *Activities) IngestComputeUrlMaps(ctx context.Context, params IngestComputeUrlMapsParams) (*IngestComputeUrlMapsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute URL map ingestion",
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
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest URL maps: %w", err))
	}

	if err := service.DeleteStaleUrlMaps(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale URL maps", "error", err)
	}

	logger.Info("Completed GCP Compute URL map ingestion",
		"projectID", params.ProjectID,
		"urlMapCount", result.UrlMapCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeUrlMapsResult{
		ProjectID:      result.ProjectID,
		UrlMapCount:    result.UrlMapCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
