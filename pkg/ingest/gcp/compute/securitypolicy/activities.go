package securitypolicy

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/gcpauth"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	entcompute "danny.vn/hotpot/pkg/storage/ent/gcp/compute"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entcompute.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entcompute.Client, limiter ratelimit.Limiter) *Activities {
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
	return NewClient(ctx, option.WithHTTPClient(httpClient))
}

// IngestComputeSecurityPoliciesParams contains parameters for the ingest activity.
type IngestComputeSecurityPoliciesParams struct {
	ProjectID string
}

// IngestComputeSecurityPoliciesResult contains the result of the ingest activity.
type IngestComputeSecurityPoliciesResult struct {
	ProjectID           string
	SecurityPolicyCount int
	DurationMillis      int64
}

// IngestComputeSecurityPoliciesActivity is the activity function reference for workflow registration.
var IngestComputeSecurityPoliciesActivity = (*Activities).IngestComputeSecurityPolicies

// IngestComputeSecurityPolicies is a Temporal activity that ingests GCP Compute security policies.
func (a *Activities) IngestComputeSecurityPolicies(ctx context.Context, params IngestComputeSecurityPoliciesParams) (*IngestComputeSecurityPoliciesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute security policy ingestion",
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
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest security policies: %w", err))
	}

	// Delete stale security policies
	if err := service.DeleteStaleSecurityPolicies(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale security policies", "error", err)
	}

	logger.Info("Completed GCP Compute security policy ingestion",
		"projectID", params.ProjectID,
		"securityPolicyCount", result.SecurityPolicyCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeSecurityPoliciesResult{
		ProjectID:           result.ProjectID,
		SecurityPolicyCount: result.SecurityPolicyCount,
		DurationMillis:      result.DurationMillis,
	}, nil
}
