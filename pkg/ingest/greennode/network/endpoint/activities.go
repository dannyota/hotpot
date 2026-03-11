package endpoint

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

// IngestNetworkEndpointsParams contains parameters for the ingest activity.
type IngestNetworkEndpointsParams struct {
	ProjectID string
	Region    string
}

// IngestNetworkEndpointsResult contains the result of the ingest activity.
type IngestNetworkEndpointsResult struct {
	EndpointCount  int
	DurationMillis int64
}

// IngestNetworkEndpointsActivity is the activity function reference for workflow registration.
var IngestNetworkEndpointsActivity = (*Activities).IngestNetworkEndpoints

// IngestNetworkEndpoints is a Temporal activity that ingests GreenNode endpoints.
func (a *Activities) IngestNetworkEndpoints(ctx context.Context, params IngestNetworkEndpointsParams) (*IngestNetworkEndpointsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode endpoint ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID, params.Region)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest endpoints: %w", err))
	}

	if err := service.DeleteStaleEndpoints(ctx, params.ProjectID, params.Region, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale endpoints", "error", err)
	}

	logger.Info("Completed GreenNode endpoint ingestion",
		"endpointCount", result.EndpointCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestNetworkEndpointsResult{
		EndpointCount:  result.EndpointCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
