package dnspolicy

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
	httpClient := &http.Client{
		Transport: ratelimit.NewRateLimitedTransport(a.limiter, nil),
	}
	return NewClient(ctx, httpClient, opts...)
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
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	// Create service
	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest policies: %w", err)
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
