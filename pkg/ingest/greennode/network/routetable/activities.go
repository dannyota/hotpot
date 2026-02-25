package routetable

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

// IngestNetworkRouteTablesParams contains parameters for the ingest activity.
type IngestNetworkRouteTablesParams struct {
	ProjectID string
	Region    string
}

// IngestNetworkRouteTablesResult contains the result of the ingest activity.
type IngestNetworkRouteTablesResult struct {
	RouteTableCount int
	DurationMillis  int64
}

// IngestNetworkRouteTablesActivity is the activity function reference for workflow registration.
var IngestNetworkRouteTablesActivity = (*Activities).IngestNetworkRouteTables

// IngestNetworkRouteTables is a Temporal activity that ingests GreenNode route tables.
func (a *Activities) IngestNetworkRouteTables(ctx context.Context, params IngestNetworkRouteTablesParams) (*IngestNetworkRouteTablesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode route table ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID, params.Region)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest route tables: %w", err))
	}

	if err := service.DeleteStaleRouteTables(ctx, params.ProjectID, params.Region, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale route tables", "error", err)
	}

	logger.Info("Completed GreenNode route table ingestion",
		"routeTableCount", result.RouteTableCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestNetworkRouteTablesResult{
		RouteTableCount: result.RouteTableCount,
		DurationMillis:  result.DurationMillis,
	}, nil
}
