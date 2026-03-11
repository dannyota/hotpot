package globalforwardingrule

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
func (a *Activities) createClient(ctx context.Context, quotaProjectID string) (*Client, error) {
	httpClient, err := gcpauth.NewHTTPClient(ctx, a.configService.GCPCredentialsJSON(), a.limiter)
	if err != nil {
		return nil, err
	}
	opts := []option.ClientOption{option.WithHTTPClient(httpClient)}
	if quotaProjectID != "" {
		opts = append(opts, option.WithQuotaProject(quotaProjectID))
	}
	return NewClient(ctx, opts...)
}

// IngestComputeGlobalForwardingRulesParams contains parameters for the ingest activity.
type IngestComputeGlobalForwardingRulesParams struct {
	ProjectID      string
	QuotaProjectID string
}

// IngestComputeGlobalForwardingRulesResult contains the result of the ingest activity.
type IngestComputeGlobalForwardingRulesResult struct {
	ProjectID                 string
	GlobalForwardingRuleCount int
	DurationMillis            int64
}

// IngestComputeGlobalForwardingRulesActivity is the activity function reference for workflow registration.
var IngestComputeGlobalForwardingRulesActivity = (*Activities).IngestComputeGlobalForwardingRules

// IngestComputeGlobalForwardingRules is a Temporal activity that ingests GCP Compute global forwarding rules.
func (a *Activities) IngestComputeGlobalForwardingRules(ctx context.Context, params IngestComputeGlobalForwardingRulesParams) (*IngestComputeGlobalForwardingRulesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute global forwarding rule ingestion",
		"projectID", params.ProjectID,
	)

	// Create client for this activity
	client, err := a.createClient(ctx, params.QuotaProjectID)
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
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest global forwarding rules: %w", err))
	}

	// Delete stale global forwarding rules
	if err := service.DeleteStaleGlobalForwardingRules(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale global forwarding rules", "error", err)
	}

	logger.Info("Completed GCP Compute global forwarding rule ingestion",
		"projectID", params.ProjectID,
		"globalForwardingRuleCount", result.GlobalForwardingRuleCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeGlobalForwardingRulesResult{
		ProjectID:                 result.ProjectID,
		GlobalForwardingRuleCount: result.GlobalForwardingRuleCount,
		DurationMillis:            result.DurationMillis,
	}, nil
}
