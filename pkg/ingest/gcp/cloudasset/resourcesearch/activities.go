package resourcesearch

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

// IngestResourceSearchParams contains parameters for the ingest activity.
type IngestResourceSearchParams struct {
}

// IngestResourceSearchResult contains the result of the ingest activity.
type IngestResourceSearchResult struct {
	ResourceCount  int
	DurationMillis int64
}

// IngestResourceSearchActivity is the activity function reference for workflow registration.
var IngestResourceSearchActivity = (*Activities).IngestResourceSearch

// IngestResourceSearch is a Temporal activity that ingests resource search results.
func (a *Activities) IngestResourceSearch(ctx context.Context, params IngestResourceSearchParams) (*IngestResourceSearchResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting Cloud Asset resource search ingestion")

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ingest resource search results: %w", err)
	}

	if err := service.DeleteStaleResources(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale resource search results", "error", err)
	}

	logger.Info("Completed Cloud Asset resource search ingestion",
		"resourceCount", result.ResourceCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestResourceSearchResult{
		ResourceCount:  result.ResourceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
