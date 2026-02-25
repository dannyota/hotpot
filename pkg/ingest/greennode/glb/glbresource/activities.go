package glbresource

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

// IngestGLBGlobalLoadBalancersParams contains parameters for the ingest activity.
type IngestGLBGlobalLoadBalancersParams struct {
	ProjectID string
	Region    string
}

// IngestGLBGlobalLoadBalancersResult contains the result of the ingest activity.
type IngestGLBGlobalLoadBalancersResult struct {
	GLBCount       int
	DurationMillis int64
}

// IngestGLBGlobalLoadBalancersActivity is the activity function reference for workflow registration.
var IngestGLBGlobalLoadBalancersActivity = (*Activities).IngestGLBGlobalLoadBalancers

// IngestGLBGlobalLoadBalancers is a Temporal activity that ingests GreenNode global load balancers.
func (a *Activities) IngestGLBGlobalLoadBalancers(ctx context.Context, params IngestGLBGlobalLoadBalancersParams) (*IngestGLBGlobalLoadBalancersResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode GLB ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest GLBs: %w", err))
	}

	if err := service.DeleteStaleGLBs(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale GLBs", "error", err)
	}

	logger.Info("Completed GreenNode GLB ingestion",
		"glbCount", result.GLBCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestGLBGlobalLoadBalancersResult{
		GLBCount:       result.GLBCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
