package interconnect

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

// IngestNetworkInterconnectsParams contains parameters for the ingest activity.
type IngestNetworkInterconnectsParams struct {
	ProjectID string
	Region    string
}

// IngestNetworkInterconnectsResult contains the result of the ingest activity.
type IngestNetworkInterconnectsResult struct {
	InterconnectCount int
	DurationMillis    int64
}

// IngestNetworkInterconnectsActivity is the activity function reference for workflow registration.
var IngestNetworkInterconnectsActivity = (*Activities).IngestNetworkInterconnects

// IngestNetworkInterconnects is a Temporal activity that ingests GreenNode interconnects.
func (a *Activities) IngestNetworkInterconnects(ctx context.Context, params IngestNetworkInterconnectsParams) (*IngestNetworkInterconnectsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode interconnect ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID, params.Region)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest interconnects: %w", err))
	}

	if err := service.DeleteStaleInterconnects(ctx, params.ProjectID, params.Region, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale interconnects", "error", err)
	}

	logger.Info("Completed GreenNode interconnect ingestion",
		"interconnectCount", result.InterconnectCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestNetworkInterconnectsResult{
		InterconnectCount: result.InterconnectCount,
		DurationMillis:    result.DurationMillis,
	}, nil
}
