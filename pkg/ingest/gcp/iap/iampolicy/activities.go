package iampolicy

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

// IngestIAPIAMPolicyParams contains parameters for the ingest activity.
type IngestIAPIAMPolicyParams struct {
	ProjectID string
}

// IngestIAPIAMPolicyResult contains the result of the ingest activity.
type IngestIAPIAMPolicyResult struct {
	ProjectID      string
	PolicyCount    int
	DurationMillis int64
}

// IngestIAPIAMPolicyActivity is the activity function reference for workflow registration.
var IngestIAPIAMPolicyActivity = (*Activities).IngestIAPIAMPolicy

// IngestIAPIAMPolicy is a Temporal activity that ingests IAP IAM policies.
func (a *Activities) IngestIAPIAMPolicy(ctx context.Context, params IngestIAPIAMPolicyParams) (*IngestIAPIAMPolicyResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting IAP IAM policy ingestion",
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
		return nil, fmt.Errorf("failed to ingest IAP IAM policy: %w", err)
	}

	if err := service.DeleteStalePolicies(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale IAP IAM policies", "error", err)
	}

	logger.Info("Completed IAP IAM policy ingestion",
		"projectID", params.ProjectID,
		"policyCount", result.PolicyCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestIAPIAMPolicyResult{
		ProjectID:      result.ProjectID,
		PolicyCount:    result.PolicyCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
