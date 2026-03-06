package vpc

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"danny.vn/greennode/auth"

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

// IngestNetworkVPCsParams contains parameters for the ingest activity.
type IngestNetworkVPCsParams struct {
	ProjectID string
	Region    string
}

// IngestNetworkVPCsResult contains the result of the ingest activity.
type IngestNetworkVPCsResult struct {
	VPCCount       int
	DurationMillis int64
}

// IngestNetworkVPCsActivity is the activity function reference for workflow registration.
var IngestNetworkVPCsActivity = (*Activities).IngestNetworkVPCs

// IngestNetworkVPCs is a Temporal activity that ingests GreenNode VPCs.
func (a *Activities) IngestNetworkVPCs(ctx context.Context, params IngestNetworkVPCsParams) (*IngestNetworkVPCsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode VPC ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID, params.Region)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest vpcs: %w", err))
	}

	if err := service.DeleteStaleVPCs(ctx, params.ProjectID, params.Region, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale vpcs", "error", err)
	}

	logger.Info("Completed GreenNode VPC ingestion",
		"vpcCount", result.VPCCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestNetworkVPCsResult{
		VPCCount:       result.VPCCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
