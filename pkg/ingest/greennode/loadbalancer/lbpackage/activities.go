package lbpackage

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"danny.vn/greennode/auth"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/base/temporalerr"
	entlb "github.com/dannyota/hotpot/pkg/storage/ent/greennode/loadbalancer"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entlb.Client
	iamAuth       *auth.IAMUserAuth
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entlb.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		iamAuth:       iamAuth,
		limiter:       limiter,
	}
}

// IngestLoadBalancerPackagesParams contains parameters for the ingest activity.
type IngestLoadBalancerPackagesParams struct {
	ProjectID string
	Region    string
}

// IngestLoadBalancerPackagesResult contains the result of the ingest activity.
type IngestLoadBalancerPackagesResult struct {
	PackageCount   int
	DurationMillis int64
}

// IngestLoadBalancerPackagesActivity is the activity function reference for workflow registration.
var IngestLoadBalancerPackagesActivity = (*Activities).IngestLoadBalancerPackages

// IngestLoadBalancerPackages is a Temporal activity that ingests GreenNode load balancer packages.
func (a *Activities) IngestLoadBalancerPackages(ctx context.Context, params IngestLoadBalancerPackagesParams) (*IngestLoadBalancerPackagesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode load balancer package ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID, params.Region)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest packages: %w", err))
	}

	if err := service.DeleteStalePackages(ctx, params.ProjectID, params.Region, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale packages", "error", err)
	}

	logger.Info("Completed GreenNode load balancer package ingestion",
		"packageCount", result.PackageCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestLoadBalancerPackagesResult{
		PackageCount:   result.PackageCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
