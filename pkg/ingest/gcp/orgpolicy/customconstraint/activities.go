package customconstraint

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
	opts = append(opts, option.WithGRPCDialOption(
		grpc.WithUnaryInterceptor(ratelimit.UnaryInterceptor(a.limiter)),
	))
	return NewClient(ctx, a.entClient, opts...)
}

type IngestCustomConstraintsParams struct {
}

type IngestCustomConstraintsResult struct {
	CustomConstraintCount int
	DurationMillis        int64
}

var IngestCustomConstraintsActivity = (*Activities).IngestCustomConstraints

func (a *Activities) IngestCustomConstraints(ctx context.Context, params IngestCustomConstraintsParams) (*IngestCustomConstraintsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting org policy custom constraint ingestion")

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ingest org policy custom constraints: %w", err)
	}

	if err := service.DeleteStaleCustomConstraints(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale org policy custom constraints", "error", err)
	}

	logger.Info("Completed org policy custom constraint ingestion",
		"customConstraintCount", result.CustomConstraintCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestCustomConstraintsResult{
		CustomConstraintCount: result.CustomConstraintCount,
		DurationMillis:        result.DurationMillis,
	}, nil
}
