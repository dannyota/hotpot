package dnspolicy

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/gcpauth"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	entdns "danny.vn/hotpot/pkg/storage/ent/gcp/dns"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entdns.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entdns.Client, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		limiter:       limiter,
	}
}

// createClient creates a rate-limited GCP client with credentials.
func (a *Activities) createClient(ctx context.Context) (*Client, error) {
	httpClient, err := gcpauth.NewHTTPClient(ctx, a.configService.GCPCredentialsJSON(), a.limiter)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, httpClient)
}

// IngestDNSPoliciesParams contains parameters for the ingest activity.
type IngestDNSPoliciesParams struct {
	ProjectID string
}

// IngestDNSPoliciesResult contains the result of the ingest activity.
type IngestDNSPoliciesResult struct {
	ProjectID      string
	PolicyCount    int
	DurationMillis int64
}

// IngestDNSPoliciesActivity is the activity function reference for workflow registration.
var IngestDNSPoliciesActivity = (*Activities).IngestDNSPolicies

// IngestDNSPolicies is a Temporal activity that ingests GCP DNS policies.
func (a *Activities) IngestDNSPolicies(ctx context.Context, params IngestDNSPoliciesParams) (*IngestDNSPoliciesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP DNS policy ingestion",
		"projectID", params.ProjectID,
	)

	// Create client for this activity
	client, err := a.createClient(ctx)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}
	defer client.Close()

	// Create service
	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest policies: %w", err))
	}

	// Delete stale policies
	if err := service.DeleteStalePolicies(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale policies", "error", err)
	}

	logger.Info("Completed GCP DNS policy ingestion",
		"projectID", params.ProjectID,
		"policyCount", result.PolicyCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestDNSPoliciesResult{
		ProjectID:      result.ProjectID,
		PolicyCount:    result.PolicyCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
