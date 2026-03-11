package healthcheck

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
func (a *Activities) createClient(ctx context.Context, quotaProjectID string) (*Client, error) {
	httpClient, err := gcpauth.NewHTTPClient(ctx, a.configService.GCPCredentialsJSON(), a.limiter)
	if err != nil {
		return nil, err
	}
	var opts []option.ClientOption
	opts = append(opts, option.WithHTTPClient(httpClient))
	if quotaProjectID != "" {
		opts = append(opts, option.WithQuotaProject(quotaProjectID))
	}
	return NewClient(ctx, opts...)
}

// IngestComputeHealthChecksParams contains parameters for the ingest activity.
type IngestComputeHealthChecksParams struct {
	ProjectID      string
	QuotaProjectID string
}

// IngestComputeHealthChecksResult contains the result of the ingest activity.
type IngestComputeHealthChecksResult struct {
	ProjectID        string
	HealthCheckCount int
	DurationMillis   int64
}

// IngestComputeHealthChecksActivity is the activity function reference for workflow registration.
var IngestComputeHealthChecksActivity = (*Activities).IngestComputeHealthChecks

// IngestComputeHealthChecks is a Temporal activity that ingests GCP Compute health checks.
func (a *Activities) IngestComputeHealthChecks(ctx context.Context, params IngestComputeHealthChecksParams) (*IngestComputeHealthChecksResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute health check ingestion",
		"projectID", params.ProjectID,
	)

	// Create client for this activity
	client, err := a.createClient(ctx, params.QuotaProjectID)
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
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest health checks: %w", err))
	}

	// Delete stale health checks
	if err := service.DeleteStaleHealthChecks(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale health checks", "error", err)
	}

	logger.Info("Completed GCP Compute health check ingestion",
		"projectID", params.ProjectID,
		"healthCheckCount", result.HealthCheckCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeHealthChecksResult{
		ProjectID:        result.ProjectID,
		HealthCheckCount: result.HealthCheckCount,
		DurationMillis:   result.DurationMillis,
	}, nil
}
