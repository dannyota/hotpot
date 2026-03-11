package neg

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

type Activities struct {
	configService *config.Service
	entClient     *entcompute.Client
	limiter       ratelimit.Limiter
}

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

type IngestComputeNegsParams struct {
	ProjectID      string
	QuotaProjectID string
}

type IngestComputeNegsResult struct {
	ProjectID      string
	NegCount       int
	DurationMillis int64
}

var IngestComputeNegsActivity = (*Activities).IngestComputeNegs

func (a *Activities) IngestComputeNegs(ctx context.Context, params IngestComputeNegsParams) (*IngestComputeNegsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute NEG ingestion", "projectID", params.ProjectID)

	client, err := a.createClient(ctx, params.QuotaProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{ProjectID: params.ProjectID})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest NEGs: %w", err))
	}

	if err := service.DeleteStaleNegs(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale NEGs", "error", err)
	}

	logger.Info("Completed GCP Compute NEG ingestion",
		"projectID", params.ProjectID,
		"negCount", result.NegCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeNegsResult{
		ProjectID:      result.ProjectID,
		NegCount:       result.NegCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
