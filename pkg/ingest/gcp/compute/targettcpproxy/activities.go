package targettcpproxy

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

// IngestComputeTargetTcpProxiesParams contains parameters for the ingest activity.
type IngestComputeTargetTcpProxiesParams struct {
	ProjectID      string
	QuotaProjectID string
}

// IngestComputeTargetTcpProxiesResult contains the result of the ingest activity.
type IngestComputeTargetTcpProxiesResult struct {
	ProjectID           string
	TargetTcpProxyCount int
	DurationMillis      int64
}

// IngestComputeTargetTcpProxiesActivity is the activity function reference for workflow registration.
var IngestComputeTargetTcpProxiesActivity = (*Activities).IngestComputeTargetTcpProxies

// IngestComputeTargetTcpProxies is a Temporal activity that ingests GCP Compute target TCP proxies.
func (a *Activities) IngestComputeTargetTcpProxies(ctx context.Context, params IngestComputeTargetTcpProxiesParams) (*IngestComputeTargetTcpProxiesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute target TCP proxy ingestion",
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
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest target TCP proxies: %w", err))
	}

	if err := service.DeleteStaleTargetTcpProxies(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale target TCP proxies", "error", err)
	}

	logger.Info("Completed GCP Compute target TCP proxy ingestion",
		"projectID", params.ProjectID,
		"targetTcpProxyCount", result.TargetTcpProxyCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeTargetTcpProxiesResult{
		ProjectID:           result.ProjectID,
		TargetTcpProxyCount: result.TargetTcpProxyCount,
		DurationMillis:      result.DurationMillis,
	}, nil
}
