package lb

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"danny.vn/greennode/auth"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
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

// IngestLoadBalancerLBsParams contains parameters for the ingest activity.
type IngestLoadBalancerLBsParams struct {
	ProjectID string
	Region    string
}

// IngestLoadBalancerLBsResult contains the result of the ingest activity.
type IngestLoadBalancerLBsResult struct {
	LBCount        int
	DurationMillis int64
}

// IngestLoadBalancerLBsActivity is the activity function reference for workflow registration.
var IngestLoadBalancerLBsActivity = (*Activities).IngestLoadBalancerLBs

// IngestLoadBalancerLBs is a Temporal activity that ingests GreenNode load balancers.
func (a *Activities) IngestLoadBalancerLBs(ctx context.Context, params IngestLoadBalancerLBsParams) (*IngestLoadBalancerLBsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode load balancer ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID, params.Region)
	if err != nil {
		return nil, fmt.Errorf("ingest load balancers: %w", err)
	}

	if err := service.DeleteStaleLBs(ctx, params.ProjectID, params.Region, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale load balancers", "error", err)
	}

	logger.Info("Completed GreenNode load balancer ingestion",
		"lbCount", result.LBCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestLoadBalancerLBsResult{
		LBCount:        result.LBCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
