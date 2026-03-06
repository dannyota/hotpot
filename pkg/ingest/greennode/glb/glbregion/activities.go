package glbregion

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"danny.vn/greennode/auth"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	entglb "danny.vn/hotpot/pkg/storage/ent/greennode/glb"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entglb.Client
	iamAuth       *auth.IAMUserAuth
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entglb.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		iamAuth:       iamAuth,
		limiter:       limiter,
	}
}

// IngestGLBGlobalRegionsParams contains parameters for the ingest activity.
type IngestGLBGlobalRegionsParams struct {
	ProjectID string
	Region    string
}

// IngestGLBGlobalRegionsResult contains the result of the ingest activity.
type IngestGLBGlobalRegionsResult struct {
	RegionCount    int
	DurationMillis int64
}

// IngestGLBGlobalRegionsActivity is the activity function reference for workflow registration.
var IngestGLBGlobalRegionsActivity = (*Activities).IngestGLBGlobalRegions

// IngestGLBGlobalRegions is a Temporal activity that ingests GreenNode global regions.
func (a *Activities) IngestGLBGlobalRegions(ctx context.Context, params IngestGLBGlobalRegionsParams) (*IngestGLBGlobalRegionsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode global region ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest global regions: %w", err))
	}

	if err := service.DeleteStaleRegions(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale global regions", "error", err)
	}

	logger.Info("Completed GreenNode global region ingestion",
		"regionCount", result.RegionCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestGLBGlobalRegionsResult{
		RegionCount:    result.RegionCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
