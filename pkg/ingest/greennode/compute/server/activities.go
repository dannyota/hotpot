package server

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"danny.vn/greennode/auth"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/base/temporalerr"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *ent.Client
	iamAuth       *auth.IAMUserAuth
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *ent.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		iamAuth:       iamAuth,
		limiter:       limiter,
	}
}

// IngestComputeServersParams contains parameters for the ingest activity.
type IngestComputeServersParams struct {
	ProjectID string
	Region    string
}

// IngestComputeServersResult contains the result of the ingest activity.
type IngestComputeServersResult struct {
	ServerCount    int
	DurationMillis int64
}

// IngestComputeServersActivity is the activity function reference for workflow registration.
var IngestComputeServersActivity = (*Activities).IngestComputeServers

// IngestComputeServers is a Temporal activity that ingests GreenNode servers.
func (a *Activities) IngestComputeServers(ctx context.Context, params IngestComputeServersParams) (*IngestComputeServersResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode server ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID, params.Region)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest servers: %w", err))
	}

	if err := service.DeleteStaleServers(ctx, params.ProjectID, params.Region, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale servers", "error", err)
	}

	logger.Info("Completed GreenNode server ingestion",
		"serverCount", result.ServerCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeServersResult{
		ServerCount:    result.ServerCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
