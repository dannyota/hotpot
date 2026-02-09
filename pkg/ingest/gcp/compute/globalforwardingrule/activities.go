package globalforwardingrule

import (
	"context"
	"fmt"
	"net/http"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *ent.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		limiter:       limiter,
	}
}

// createClient creates a rate-limited GCP client with credentials.
func (a *Activities) createClient(ctx context.Context) (*Client, error) {
	var opts []option.ClientOption
	if credJSON := a.configService.GCPCredentialsJSON(); len(credJSON) > 0 {
		opts = append(opts, option.WithAuthCredentialsJSON(option.ServiceAccount, credJSON))
	}
	opts = append(opts, option.WithHTTPClient(&http.Client{
		Transport: ratelimit.NewRateLimitedTransport(a.limiter, nil),
	}))
	return NewClient(ctx, opts...)
}

// IngestComputeGlobalForwardingRulesParams contains parameters for the ingest activity.
type IngestComputeGlobalForwardingRulesParams struct {
	ProjectID string
}

// IngestComputeGlobalForwardingRulesResult contains the result of the ingest activity.
type IngestComputeGlobalForwardingRulesResult struct {
	ProjectID                string
	GlobalForwardingRuleCount int
	DurationMillis           int64
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
	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	// Create service
	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest global forwarding rules: %w", err)
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
		ProjectID:                result.ProjectID,
		GlobalForwardingRuleCount: result.GlobalForwardingRuleCount,
		DurationMillis:           result.DurationMillis,
	}, nil
}
