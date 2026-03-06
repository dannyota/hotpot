package servergroup

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

// IngestComputeServerGroupsParams contains parameters for the ingest activity.
type IngestComputeServerGroupsParams struct {
	ProjectID string
	Region    string
}

// IngestComputeServerGroupsResult contains the result of the ingest activity.
type IngestComputeServerGroupsResult struct {
	GroupCount     int
	DurationMillis int64
}

// IngestComputeServerGroupsActivity is the activity function reference for workflow registration.
var IngestComputeServerGroupsActivity = (*Activities).IngestComputeServerGroups

// IngestComputeServerGroups is a Temporal activity that ingests GreenNode server groups.
func (a *Activities) IngestComputeServerGroups(ctx context.Context, params IngestComputeServerGroupsParams) (*IngestComputeServerGroupsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode server group ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID, params.Region)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest server groups: %w", err))
	}

	if err := service.DeleteStaleServerGroups(ctx, params.ProjectID, params.Region, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale server groups", "error", err)
	}

	logger.Info("Completed GreenNode server group ingestion",
		"groupCount", result.GroupCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeServerGroupsResult{
		GroupCount:     result.GroupCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
