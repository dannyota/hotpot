package secgroup

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

// IngestNetworkSecgroupsParams contains parameters for the ingest activity.
type IngestNetworkSecgroupsParams struct {
	ProjectID string
	Region    string
}

// IngestNetworkSecgroupsResult contains the result of the ingest activity.
type IngestNetworkSecgroupsResult struct {
	SecgroupCount  int
	DurationMillis int64
}

// IngestNetworkSecgroupsActivity is the activity function reference for workflow registration.
var IngestNetworkSecgroupsActivity = (*Activities).IngestNetworkSecgroups

// IngestNetworkSecgroups is a Temporal activity that ingests GreenNode security groups.
func (a *Activities) IngestNetworkSecgroups(ctx context.Context, params IngestNetworkSecgroupsParams) (*IngestNetworkSecgroupsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode security group ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID, params.Region)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest secgroups: %w", err))
	}

	if err := service.DeleteStaleSecgroups(ctx, params.ProjectID, params.Region, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale secgroups", "error", err)
	}

	logger.Info("Completed GreenNode security group ingestion",
		"secgroupCount", result.SecgroupCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestNetworkSecgroupsResult{
		SecgroupCount:  result.SecgroupCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
