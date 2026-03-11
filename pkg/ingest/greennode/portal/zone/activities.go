package zone

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"danny.vn/gnode/auth"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	entportal "danny.vn/hotpot/pkg/storage/ent/greennode/portal"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entportal.Client
	iamAuth       *auth.IAMUserAuth
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entportal.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		iamAuth:       iamAuth,
		limiter:       limiter,
	}
}

// IngestPortalZonesParams contains parameters for the ingest activity.
type IngestPortalZonesParams struct {
	ProjectID string
	Region    string
}

// IngestPortalZonesResult contains the result of the ingest activity.
type IngestPortalZonesResult struct {
	ZoneCount      int
	DurationMillis int64
}

// IngestPortalZonesActivity is the activity function reference for workflow registration.
var IngestPortalZonesActivity = (*Activities).IngestPortalZones

// IngestPortalZones is a Temporal activity that ingests GreenNode zones.
func (a *Activities) IngestPortalZones(ctx context.Context, params IngestPortalZonesParams) (*IngestPortalZonesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode zone ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest zones: %w", err))
	}

	if err := service.DeleteStaleZones(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale zones", "error", err)
	}

	logger.Info("Completed GreenNode zone ingestion",
		"zoneCount", result.ZoneCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestPortalZonesResult{
		ZoneCount:      result.ZoneCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
