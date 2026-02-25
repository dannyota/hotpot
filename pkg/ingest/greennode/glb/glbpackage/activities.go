package glbpackage

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"danny.vn/greennode/auth"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/base/temporalerr"
	entglb "github.com/dannyota/hotpot/pkg/storage/ent/greennode/glb"
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

// IngestGLBGlobalPackagesParams contains parameters for the ingest activity.
type IngestGLBGlobalPackagesParams struct {
	ProjectID string
	Region    string
}

// IngestGLBGlobalPackagesResult contains the result of the ingest activity.
type IngestGLBGlobalPackagesResult struct {
	PackageCount   int
	DurationMillis int64
}

// IngestGLBGlobalPackagesActivity is the activity function reference for workflow registration.
var IngestGLBGlobalPackagesActivity = (*Activities).IngestGLBGlobalPackages

// IngestGLBGlobalPackages is a Temporal activity that ingests GreenNode global packages.
func (a *Activities) IngestGLBGlobalPackages(ctx context.Context, params IngestGLBGlobalPackagesParams) (*IngestGLBGlobalPackagesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode global package ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest global packages: %w", err))
	}

	if err := service.DeleteStalePackages(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale global packages", "error", err)
	}

	logger.Info("Completed GreenNode global package ingestion",
		"packageCount", result.PackageCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestGLBGlobalPackagesResult{
		PackageCount:   result.PackageCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
