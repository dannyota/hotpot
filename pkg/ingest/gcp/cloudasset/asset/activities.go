package asset

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

// IngestAssetsParams contains parameters for the ingest activity.
type IngestAssetsParams struct {
}

// IngestAssetsResult contains the result of the ingest activity.
type IngestAssetsResult struct {
	AssetCount     int
	DurationMillis int64
}

// IngestAssetsActivity is the activity function reference for workflow registration.
var IngestAssetsActivity = (*Activities).IngestAssets

// IngestAssets is a Temporal activity that ingests Cloud Asset assets.
func (a *Activities) IngestAssets(ctx context.Context, params IngestAssetsParams) (*IngestAssetsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting Cloud Asset asset ingestion")

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ingest Cloud Asset assets: %w", err)
	}

	if err := service.DeleteStaleAssets(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale Cloud Asset assets", "error", err)
	}

	logger.Info("Completed Cloud Asset asset ingestion",
		"assetCount", result.AssetCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestAssetsResult{
		AssetCount:     result.AssetCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
