package negendpoint

import (
	"context"
	"fmt"
	"net/http"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

type Activities struct {
	configService *config.Service
	entClient     *ent.Client
	limiter       ratelimit.Limiter
}

func NewActivities(configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		limiter:       limiter,
	}
}

func (a *Activities) createClient(ctx context.Context) (*Client, error) {
	var opts []option.ClientOption
	if credJSON := a.configService.GCPCredentialsJSON(); len(credJSON) > 0 {
		opts = append(opts, option.WithAuthCredentialsJSON(option.ServiceAccount, credJSON))
	}
	opts = append(opts, option.WithHTTPClient(&http.Client{
		Transport: ratelimit.NewRateLimitedTransport(a.limiter, nil),
	}))
	return NewClient(ctx, a.entClient, opts...)
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
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{ProjectID: params.ProjectID})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest NEG endpoints: %w", err)
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
