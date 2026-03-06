package volumetypezone

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"danny.vn/greennode/auth"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	entvol "danny.vn/hotpot/pkg/storage/ent/greennode/volume"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entvol.Client
	iamAuth       *auth.IAMUserAuth
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entvol.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		iamAuth:       iamAuth,
		limiter:       limiter,
	}
}

// IngestVolumeVolumeTypeZonesParams contains parameters for the ingest activity.
type IngestVolumeVolumeTypeZonesParams struct {
	ProjectID string
	Region    string
}

// IngestVolumeVolumeTypeZonesResult contains the result of the ingest activity.
type IngestVolumeVolumeTypeZonesResult struct {
	VolumeTypeZoneCount int
	DurationMillis      int64
}

// IngestVolumeVolumeTypeZonesActivity is the activity function reference for workflow registration.
var IngestVolumeVolumeTypeZonesActivity = (*Activities).IngestVolumeVolumeTypeZones

// IngestVolumeVolumeTypeZones is a Temporal activity that ingests GreenNode volume type zones.
func (a *Activities) IngestVolumeVolumeTypeZones(ctx context.Context, params IngestVolumeVolumeTypeZonesParams) (*IngestVolumeVolumeTypeZonesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode volume type zone ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID, params.Region)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest volume type zones: %w", err))
	}

	if err := service.DeleteStaleVolumeTypeZones(ctx, params.ProjectID, params.Region, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale volume type zones", "error", err)
	}

	logger.Info("Completed GreenNode volume type zone ingestion",
		"volumeTypeZoneCount", result.VolumeTypeZoneCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestVolumeVolumeTypeZonesResult{
		VolumeTypeZoneCount: result.VolumeTypeZoneCount,
		DurationMillis:      result.DurationMillis,
	}, nil
}
