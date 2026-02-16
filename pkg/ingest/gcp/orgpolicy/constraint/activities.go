package constraint

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

type IngestConstraintsParams struct {
}

type IngestConstraintsResult struct {
	ConstraintCount int
	DurationMillis  int64
}

var IngestConstraintsActivity = (*Activities).IngestConstraints

func (a *Activities) IngestConstraints(ctx context.Context, params IngestConstraintsParams) (*IngestConstraintsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting org policy constraint ingestion")

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ingest org policy constraints: %w", err)
	}

	if err := service.DeleteStaleConstraints(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale org policy constraints", "error", err)
	}

	logger.Info("Completed org policy constraint ingestion",
		"constraintCount", result.ConstraintCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestConstraintsResult{
		ConstraintCount: result.ConstraintCount,
		DurationMillis:  result.DurationMillis,
	}, nil
}
