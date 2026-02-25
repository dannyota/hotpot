package region

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

// IngestPortalRegionsParams contains parameters for the ingest activity.
type IngestPortalRegionsParams struct {
	ProjectID string
	Region    string
}

// IngestPortalRegionsResult contains the result of the ingest activity.
type IngestPortalRegionsResult struct {
	RegionCount    int
	DurationMillis int64
}

// IngestPortalRegionsActivity is the activity function reference for workflow registration.
var IngestPortalRegionsActivity = (*Activities).IngestPortalRegions

// IngestPortalRegions is a Temporal activity that ingests GreenNode regions.
func (a *Activities) IngestPortalRegions(ctx context.Context, params IngestPortalRegionsParams) (*IngestPortalRegionsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode region ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest regions: %w", err))
	}

	if err := service.DeleteStaleRegions(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale regions", "error", err)
	}

	logger.Info("Completed GreenNode region ingestion",
		"regionCount", result.RegionCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestPortalRegionsResult{
		RegionCount:    result.RegionCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
