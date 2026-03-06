package peering

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

// IngestNetworkPeeringsParams contains parameters for the ingest activity.
type IngestNetworkPeeringsParams struct {
	ProjectID string
	Region    string
}

// IngestNetworkPeeringsResult contains the result of the ingest activity.
type IngestNetworkPeeringsResult struct {
	PeeringCount   int
	DurationMillis int64
}

// IngestNetworkPeeringsActivity is the activity function reference for workflow registration.
var IngestNetworkPeeringsActivity = (*Activities).IngestNetworkPeerings

// IngestNetworkPeerings is a Temporal activity that ingests GreenNode peerings.
func (a *Activities) IngestNetworkPeerings(ctx context.Context, params IngestNetworkPeeringsParams) (*IngestNetworkPeeringsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode peering ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID, params.Region)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest peerings: %w", err))
	}

	if err := service.DeleteStalePeerings(ctx, params.ProjectID, params.Region, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale peerings", "error", err)
	}

	logger.Info("Completed GreenNode peering ingestion",
		"peeringCount", result.PeeringCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestNetworkPeeringsResult{
		PeeringCount:   result.PeeringCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
