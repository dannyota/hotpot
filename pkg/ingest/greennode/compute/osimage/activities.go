package osimage

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"danny.vn/greennode/auth"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/base/temporalerr"
	entcompute "github.com/dannyota/hotpot/pkg/storage/ent/greennode/compute"
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

// IngestComputeOSImagesParams contains parameters for the ingest activity.
type IngestComputeOSImagesParams struct {
	ProjectID string
	Region    string
}

// IngestComputeOSImagesResult contains the result of the ingest activity.
type IngestComputeOSImagesResult struct {
	OSImageCount   int
	DurationMillis int64
}

// IngestComputeOSImagesActivity is the activity function reference for workflow registration.
var IngestComputeOSImagesActivity = (*Activities).IngestComputeOSImages

// IngestComputeOSImages is a Temporal activity that ingests GreenNode OS images.
func (a *Activities) IngestComputeOSImages(ctx context.Context, params IngestComputeOSImagesParams) (*IngestComputeOSImagesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode OS image ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID, params.Region)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest os images: %w", err))
	}

	if err := service.DeleteStaleOSImages(ctx, params.ProjectID, params.Region, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale os images", "error", err)
	}

	logger.Info("Completed GreenNode OS image ingestion",
		"osImageCount", result.OSImageCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeOSImagesResult{
		OSImageCount:   result.OSImageCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
