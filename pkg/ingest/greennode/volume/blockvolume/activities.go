package blockvolume

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

// IngestVolumeBlockVolumesParams contains parameters for the ingest activity.
type IngestVolumeBlockVolumesParams struct {
	ProjectID string
	Region    string
}

// IngestVolumeBlockVolumesResult contains the result of the ingest activity.
type IngestVolumeBlockVolumesResult struct {
	BlockVolumeCount int
	DurationMillis   int64
}

// IngestVolumeBlockVolumesActivity is the activity function reference for workflow registration.
var IngestVolumeBlockVolumesActivity = (*Activities).IngestVolumeBlockVolumes

// IngestVolumeBlockVolumes is a Temporal activity that ingests GreenNode block volumes.
func (a *Activities) IngestVolumeBlockVolumes(ctx context.Context, params IngestVolumeBlockVolumesParams) (*IngestVolumeBlockVolumesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode block volume ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID, params.Region)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest block volumes: %w", err))
	}

	if err := service.DeleteStaleBlockVolumes(ctx, params.ProjectID, params.Region, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale block volumes", "error", err)
	}

	logger.Info("Completed GreenNode block volume ingestion",
		"blockVolumeCount", result.BlockVolumeCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestVolumeBlockVolumesResult{
		BlockVolumeCount: result.BlockVolumeCount,
		DurationMillis:   result.DurationMillis,
	}, nil
}
