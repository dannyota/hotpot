package negendpoint

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
func (a *Activities) createClient(ctx context.Context) (*Client, error) {
	httpClient, err := gcpauth.NewHTTPClient(ctx, a.configService.GCPCredentialsJSON(), a.limiter)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, a.entClient, option.WithHTTPClient(httpClient))
}

type IngestComputeNegEndpointsParams struct {
	ProjectID string
}

type IngestComputeNegEndpointsResult struct {
	ProjectID        string
	NegEndpointCount int
	DurationMillis   int64
}

var IngestComputeNegEndpointsActivity = (*Activities).IngestComputeNegEndpoints

func (a *Activities) IngestComputeNegEndpoints(ctx context.Context, params IngestComputeNegEndpointsParams) (*IngestComputeNegEndpointsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute NEG endpoint ingestion", "projectID", params.ProjectID)

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{ProjectID: params.ProjectID})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest NEG endpoints: %w", err))
	}

	if err := service.DeleteStaleNegEndpoints(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale NEG endpoints", "error", err)
	}

	logger.Info("Completed GCP Compute NEG endpoint ingestion",
		"projectID", params.ProjectID,
		"negEndpointCount", result.NegEndpointCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeNegEndpointsResult{
		ProjectID:        result.ProjectID,
		NegEndpointCount: result.NegEndpointCount,
		DurationMillis:   result.DurationMillis,
	}, nil
}
