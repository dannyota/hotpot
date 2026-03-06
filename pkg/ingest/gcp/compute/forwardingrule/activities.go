package forwardingrule

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

// IngestComputeForwardingRulesParams contains parameters for the ingest activity.
type IngestComputeForwardingRulesParams struct {
	ProjectID string
}

// IngestComputeForwardingRulesResult contains the result of the ingest activity.
type IngestComputeForwardingRulesResult struct {
	ProjectID           string
	ForwardingRuleCount int
	DurationMillis      int64
}

// IngestComputeForwardingRulesActivity is the activity function reference for workflow registration.
var IngestComputeForwardingRulesActivity = (*Activities).IngestComputeForwardingRules

// IngestComputeForwardingRules is a Temporal activity that ingests GCP Compute regional forwarding rules.
func (a *Activities) IngestComputeForwardingRules(ctx context.Context, params IngestComputeForwardingRulesParams) (*IngestComputeForwardingRulesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute forwarding rule ingestion",
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
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest forwarding rules: %w", err))
	}

	// Delete stale forwarding rules
	if err := service.DeleteStaleForwardingRules(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale forwarding rules", "error", err)
	}

	logger.Info("Completed GCP Compute forwarding rule ingestion",
		"projectID", params.ProjectID,
		"forwardingRuleCount", result.ForwardingRuleCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeForwardingRulesResult{
		ProjectID:           result.ProjectID,
		ForwardingRuleCount: result.ForwardingRuleCount,
		DurationMillis:      result.DurationMillis,
	}, nil
}
