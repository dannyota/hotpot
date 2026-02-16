package alertpolicy

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

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
	opts = append(opts, option.WithGRPCDialOption(
		grpc.WithUnaryInterceptor(ratelimit.UnaryInterceptor(a.limiter)),
	))
	return NewClient(ctx, opts...)
}

// IngestAlertPoliciesParams contains parameters for the ingest activity.
type IngestAlertPoliciesParams struct {
	ProjectID string
}

// IngestAlertPoliciesResult contains the result of the ingest activity.
type IngestAlertPoliciesResult struct {
	ProjectID        string
	AlertPolicyCount int
	DurationMillis   int64
}

// IngestAlertPoliciesActivity is the activity function reference for workflow registration.
var IngestAlertPoliciesActivity = (*Activities).IngestAlertPolicies

// IngestAlertPolicies is a Temporal activity that ingests Monitoring alert policies.
func (a *Activities) IngestAlertPolicies(ctx context.Context, params IngestAlertPoliciesParams) (*IngestAlertPoliciesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting Monitoring alert policy ingestion",
		"projectID", params.ProjectID,
	)

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest alert policies: %w", err)
	}

	if err := service.DeleteStaleAlertPolicies(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale alert policies", "error", err)
	}

	logger.Info("Completed Monitoring alert policy ingestion",
		"projectID", params.ProjectID,
		"alertPolicyCount", result.AlertPolicyCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestAlertPoliciesResult{
		ProjectID:        result.ProjectID,
		AlertPolicyCount: result.AlertPolicyCount,
		DurationMillis:   result.DurationMillis,
	}, nil
}
