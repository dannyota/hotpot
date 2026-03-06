package certificate

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"danny.vn/greennode/auth"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	entlb "danny.vn/hotpot/pkg/storage/ent/greennode/loadbalancer"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entlb.Client
	iamAuth       *auth.IAMUserAuth
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entlb.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		iamAuth:       iamAuth,
		limiter:       limiter,
	}
}

// IngestLoadBalancerCertificatesParams contains parameters for the ingest activity.
type IngestLoadBalancerCertificatesParams struct {
	ProjectID string
	Region    string
}

// IngestLoadBalancerCertificatesResult contains the result of the ingest activity.
type IngestLoadBalancerCertificatesResult struct {
	CertificateCount int
	DurationMillis   int64
}

// IngestLoadBalancerCertificatesActivity is the activity function reference for workflow registration.
var IngestLoadBalancerCertificatesActivity = (*Activities).IngestLoadBalancerCertificates

// IngestLoadBalancerCertificates is a Temporal activity that ingests GreenNode certificates.
func (a *Activities) IngestLoadBalancerCertificates(ctx context.Context, params IngestLoadBalancerCertificatesParams) (*IngestLoadBalancerCertificatesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GreenNode certificate ingestion", "projectID", params.ProjectID, "region", params.Region)

	client, err := NewClient(ctx, a.configService, a.iamAuth, a.limiter, params.Region, params.ProjectID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, params.ProjectID, params.Region)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest certificates: %w", err))
	}

	if err := service.DeleteStaleCertificates(ctx, params.ProjectID, params.Region, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale certificates", "error", err)
	}

	logger.Info("Completed GreenNode certificate ingestion",
		"certificateCount", result.CertificateCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestLoadBalancerCertificatesResult{
		CertificateCount: result.CertificateCount,
		DurationMillis:   result.DurationMillis,
	}, nil
}
