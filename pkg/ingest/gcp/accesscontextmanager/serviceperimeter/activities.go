package serviceperimeter

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

// IngestServicePerimetersParams contains parameters for the ingest activity.
type IngestServicePerimetersParams struct {
}

// IngestServicePerimetersResult contains the result of the ingest activity.
type IngestServicePerimetersResult struct {
	PerimeterCount int
	DurationMillis int64
}

// IngestServicePerimetersActivity is the activity function reference for workflow registration.
var IngestServicePerimetersActivity = (*Activities).IngestServicePerimeters

// IngestServicePerimeters is a Temporal activity that ingests service perimeters.
func (a *Activities) IngestServicePerimeters(ctx context.Context, params IngestServicePerimetersParams) (*IngestServicePerimetersResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting service perimeter ingestion")

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ingest service perimeters: %w", err)
	}

	if err := service.DeleteStalePerimeters(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale service perimeters", "error", err)
	}

	logger.Info("Completed service perimeter ingestion",
		"perimeterCount", result.PerimeterCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestServicePerimetersResult{
		PerimeterCount: result.PerimeterCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
