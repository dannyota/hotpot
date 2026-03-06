package userimage

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"danny.vn/greennode/auth"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	entcompute "danny.vn/hotpot/pkg/storage/ent/greennode/compute"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entcompute.Client
	iamAuth       *auth.IAMUserAuth
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entcompute.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		iamAuth:       iamAuth,
		limiter:       limiter,
	}
}

// IngestComputeUserImagesParams contains parameters for the ingest activity.
type IngestComputeUserImagesParams struct {
	ProjectID string
	Region    string
}

// IngestComputeUserImagesResult contains the result of the ingest activity.
type IngestComputeUserImagesResult struct {
	UserImageCount int
	DurationMillis int64
}

// IngestComputeUserImagesActivity is the activity function reference for workflow registration.
var IngestComputeUserImagesActivity = (*Activities).IngestComputeUserImages

// IngestComputeUserImages is a Temporal activity that ingests GreenNode user images.
func (a *Activities) IngestComputeUserImages(ctx context.Context, params IngestComputeUserImagesParams) (*IngestComputeUserImagesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode user image ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID, params.Region)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest user images: %w", err))
	}

	if err := service.DeleteStaleUserImages(ctx, params.ProjectID, params.Region, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale user images", "error", err)
	}

	logger.Info("Completed GreenNode user image ingestion",
		"userImageCount", result.UserImageCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeUserImagesResult{
		UserImageCount: result.UserImageCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
