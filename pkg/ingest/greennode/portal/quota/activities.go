package quota

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"danny.vn/gnode/auth"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	entportal "danny.vn/hotpot/pkg/storage/ent/greennode/portal"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entportal.Client
	iamAuth       *auth.IAMUserAuth
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entportal.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		iamAuth:       iamAuth,
		limiter:       limiter,
	}
}

// IngestPortalQuotasParams contains parameters for the ingest activity.
type IngestPortalQuotasParams struct {
	ProjectID string
	Region    string
}

// IngestPortalQuotasResult contains the result of the ingest activity.
type IngestPortalQuotasResult struct {
	QuotaCount     int
	DurationMillis int64
}

// IngestPortalQuotasActivity is the activity function reference for workflow registration.
var IngestPortalQuotasActivity = (*Activities).IngestPortalQuotas

// IngestPortalQuotas is a Temporal activity that ingests GreenNode quotas.
func (a *Activities) IngestPortalQuotas(ctx context.Context, params IngestPortalQuotasParams) (*IngestPortalQuotasResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode quota ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID, params.Region)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest quotas: %w", err))
	}

	if err := service.DeleteStaleQuotas(ctx, params.ProjectID, params.Region, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale quotas", "error", err)
	}

	logger.Info("Completed GreenNode quota ingestion",
		"quotaCount", result.QuotaCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestPortalQuotasResult{
		QuotaCount:     result.QuotaCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
