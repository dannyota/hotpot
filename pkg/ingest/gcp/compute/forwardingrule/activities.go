package forwardingrule

import (
	"context"
	"fmt"
	"net/http"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	db            *gorm.DB
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, db *gorm.DB, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		db:            db,
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

// IngestComputeForwardingRulesParams contains parameters for the ingest activity.
type IngestComputeForwardingRulesParams struct {
	ProjectID string
}

// IngestComputeForwardingRulesResult contains the result of the ingest activity.
type IngestComputeForwardingRulesResult struct {
	ProjectID          string
	ForwardingRuleCount int
	DurationMillis     int64
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
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	// Create service
	service := NewService(client, a.db)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest forwarding rules: %w", err)
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
		ProjectID:          result.ProjectID,
		ForwardingRuleCount: result.ForwardingRuleCount,
		DurationMillis:     result.DurationMillis,
	}, nil
}
