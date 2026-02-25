package hostedzone

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

// IngestDNSHostedZonesParams contains parameters for the ingest activity.
type IngestDNSHostedZonesParams struct {
	ProjectID string
	Region    string
}

// IngestDNSHostedZonesResult contains the result of the ingest activity.
type IngestDNSHostedZonesResult struct {
	HostedZoneCount int
	DurationMillis  int64
}

// IngestDNSHostedZonesActivity is the activity function reference for workflow registration.
var IngestDNSHostedZonesActivity = (*Activities).IngestDNSHostedZones

// IngestDNSHostedZones is a Temporal activity that ingests GreenNode DNS hosted zones.
func (a *Activities) IngestDNSHostedZones(ctx context.Context, params IngestDNSHostedZonesParams) (*IngestDNSHostedZonesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode DNS hosted zone ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest hosted zones: %w", err))
	}

	if err := service.DeleteStaleHostedZones(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale hosted zones", "error", err)
	}

	logger.Info("Completed GreenNode DNS hosted zone ingestion",
		"hostedZoneCount", result.HostedZoneCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestDNSHostedZonesResult{
		HostedZoneCount: result.HostedZoneCount,
		DurationMillis:  result.DurationMillis,
	}, nil
}
