package neg

import (
	"context"
	"fmt"
	"net/http"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
	"hotpot/pkg/storage/ent"
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
	return NewClient(ctx, opts...)
}

type IngestComputeNegsParams struct {
	ProjectID string
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

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{ProjectID: params.ProjectID})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest NEGs: %w", err)
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
