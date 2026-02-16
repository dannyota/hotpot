package source

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *ent.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		limiter:       limiter,
	}
}

// createClient creates a rate-limited GCP client with credentials.
func (a *Activities) createClient(ctx context.Context) (*Client, error) {
	var opts []option.ClientOption
	if credJSON := a.configService.GCPCredentialsJSON(); len(credJSON) > 0 {
		opts = append(opts, option.WithAuthCredentialsJSON(option.ServiceAccount, credJSON))
	}
	opts = append(opts, option.WithGRPCDialOption(
		grpc.WithUnaryInterceptor(ratelimit.UnaryInterceptor(a.limiter)),
	))
	return NewClient(ctx, a.entClient, opts...)
}

// IngestSourcesParams contains parameters for the ingest activity.
type IngestSourcesParams struct {
}

// IngestSourcesResult contains the result of the ingest activity.
type IngestSourcesResult struct {
	SourceCount    int
	DurationMillis int64
}

// IngestSourcesActivity is the activity function reference for workflow registration.
var IngestSourcesActivity = (*Activities).IngestSources

// IngestSources is a Temporal activity that ingests SCC sources.
func (a *Activities) IngestSources(ctx context.Context, params IngestSourcesParams) (*IngestSourcesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting SCC source ingestion")

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ingest SCC sources: %w", err)
	}

	if err := service.DeleteStaleSources(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale SCC sources", "error", err)
	}

	logger.Info("Completed SCC source ingestion",
		"sourceCount", result.SourceCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestSourcesResult{
		SourceCount:    result.SourceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
