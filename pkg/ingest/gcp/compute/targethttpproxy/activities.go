package targethttpproxy

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

// IngestComputeTargetHttpProxiesParams contains parameters for the ingest activity.
type IngestComputeTargetHttpProxiesParams struct {
	ProjectID      string
	QuotaProjectID string
}

// IngestComputeTargetHttpProxiesResult contains the result of the ingest activity.
type IngestComputeTargetHttpProxiesResult struct {
	ProjectID            string
	TargetHttpProxyCount int
	DurationMillis       int64
}

// IngestComputeTargetHttpProxiesActivity is the activity function reference for workflow registration.
var IngestComputeTargetHttpProxiesActivity = (*Activities).IngestComputeTargetHttpProxies

// IngestComputeTargetHttpProxies is a Temporal activity that ingests GCP Compute target HTTP proxies.
func (a *Activities) IngestComputeTargetHttpProxies(ctx context.Context, params IngestComputeTargetHttpProxiesParams) (*IngestComputeTargetHttpProxiesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute target HTTP proxy ingestion",
		"projectID", params.ProjectID,
	)

	client, err := a.createClient(ctx, params.QuotaProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest target HTTP proxies: %w", err))
	}

	if err := service.DeleteStaleTargetHttpProxies(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale target HTTP proxies", "error", err)
	}

	logger.Info("Completed GCP Compute target HTTP proxy ingestion",
		"projectID", params.ProjectID,
		"targetHttpProxyCount", result.TargetHttpProxyCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeTargetHttpProxiesResult{
		ProjectID:            result.ProjectID,
		TargetHttpProxyCount: result.TargetHttpProxyCount,
		DurationMillis:       result.DurationMillis,
	}, nil
}
