package accesslevel

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

// IngestAccessLevelsParams contains parameters for the ingest activity.
type IngestAccessLevelsParams struct {
}

// IngestAccessLevelsResult contains the result of the ingest activity.
type IngestAccessLevelsResult struct {
	LevelCount     int
	DurationMillis int64
}

// IngestAccessLevelsActivity is the activity function reference for workflow registration.
var IngestAccessLevelsActivity = (*Activities).IngestAccessLevels

// IngestAccessLevels is a Temporal activity that ingests access levels.
func (a *Activities) IngestAccessLevels(ctx context.Context, params IngestAccessLevelsParams) (*IngestAccessLevelsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting access level ingestion")

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ingest access levels: %w", err)
	}

	if err := service.DeleteStaleLevels(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale access levels", "error", err)
	}

	logger.Info("Completed access level ingestion",
		"levelCount", result.LevelCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestAccessLevelsResult{
		LevelCount:     result.LevelCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
