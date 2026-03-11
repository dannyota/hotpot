package accesspolicy

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	entaccesscontextmanager "danny.vn/hotpot/pkg/storage/ent/gcp/accesscontextmanager"
	entresourcemanager "danny.vn/hotpot/pkg/storage/ent/gcp/resourcemanager"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entaccesscontextmanager.Client
	rmClient      *entresourcemanager.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entaccesscontextmanager.Client, rmClient *entresourcemanager.Client, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		rmClient:      rmClient,
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
	return NewClient(ctx, a.entClient, a.rmClient, opts...)
}

// IngestAccessPoliciesParams contains parameters for the ingest activity.
type IngestAccessPoliciesParams struct {
}

// IngestAccessPoliciesResult contains the result of the ingest activity.
type IngestAccessPoliciesResult struct {
	PolicyCount    int
	DurationMillis int64
}

// IngestAccessPoliciesActivity is the activity function reference for workflow registration.
var IngestAccessPoliciesActivity = (*Activities).IngestAccessPolicies

// IngestAccessPolicies is a Temporal activity that ingests access policies.
func (a *Activities) IngestAccessPolicies(ctx context.Context, params IngestAccessPoliciesParams) (*IngestAccessPoliciesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting access policy ingestion")

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("create client: %w", err))
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest access policies: %w", err))
	}

	if err := service.DeleteStalePolicies(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale access policies", "error", err)
	}

	logger.Info("Completed access policy ingestion",
		"policyCount", result.PolicyCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestAccessPoliciesResult{
		PolicyCount:    result.PolicyCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
