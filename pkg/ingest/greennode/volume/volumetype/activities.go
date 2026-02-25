package volumetype

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"danny.vn/greennode/auth"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/base/temporalerr"
	entvol "github.com/dannyota/hotpot/pkg/storage/ent/greennode/volume"
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

// IngestVolumeVolumeTypesParams contains parameters for the ingest activity.
type IngestVolumeVolumeTypesParams struct {
	ProjectID string
	Region    string
}

// IngestVolumeVolumeTypesResult contains the result of the ingest activity.
type IngestVolumeVolumeTypesResult struct {
	VolumeTypeCount int
	DurationMillis  int64
}

// IngestVolumeVolumeTypesActivity is the activity function reference for workflow registration.
var IngestVolumeVolumeTypesActivity = (*Activities).IngestVolumeVolumeTypes

// IngestVolumeVolumeTypes is a Temporal activity that ingests GreenNode volume types.
func (a *Activities) IngestVolumeVolumeTypes(ctx context.Context, params IngestVolumeVolumeTypesParams) (*IngestVolumeVolumeTypesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode volume type ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID, params.Region)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest volume types: %w", err))
	}

	if err := service.DeleteStaleVolumeTypes(ctx, params.ProjectID, params.Region, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale volume types", "error", err)
	}

	logger.Info("Completed GreenNode volume type ingestion",
		"volumeTypeCount", result.VolumeTypeCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestVolumeVolumeTypesResult{
		VolumeTypeCount: result.VolumeTypeCount,
		DurationMillis:  result.DurationMillis,
	}, nil
}
