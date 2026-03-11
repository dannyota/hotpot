package subnet

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"danny.vn/gnode/auth"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	entnet "danny.vn/hotpot/pkg/storage/ent/greennode/network"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entnet.Client
	iamAuth       *auth.IAMUserAuth
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entnet.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		iamAuth:       iamAuth,
		limiter:       limiter,
	}
}

// IngestNetworkSubnetsParams contains parameters for the ingest activity.
type IngestNetworkSubnetsParams struct {
	ProjectID string
	Region    string
}

// IngestNetworkSubnetsResult contains the result of the ingest activity.
type IngestNetworkSubnetsResult struct {
	SubnetCount    int
	DurationMillis int64
}

// IngestNetworkSubnetsActivity is the activity function reference for workflow registration.
var IngestNetworkSubnetsActivity = (*Activities).IngestNetworkSubnets

// IngestNetworkSubnets is a Temporal activity that ingests GreenNode subnets.
func (a *Activities) IngestNetworkSubnets(ctx context.Context, params IngestNetworkSubnetsParams) (*IngestNetworkSubnetsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode subnet ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID, params.Region)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest subnets: %w", err))
	}

	if err := service.DeleteStaleSubnets(ctx, params.ProjectID, params.Region, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale subnets", "error", err)
	}

	logger.Info("Completed GreenNode subnet ingestion",
		"subnetCount", result.SubnetCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestNetworkSubnetsResult{
		SubnetCount:    result.SubnetCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
