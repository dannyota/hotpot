package bucketiam

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
	return NewClient(ctx, a.entClient, httpClient, opts...)
}

// IngestStorageBucketIamPoliciesParams contains parameters for the ingest activity.
type IngestStorageBucketIamPoliciesParams struct {
	ProjectID string
}

// IngestStorageBucketIamPoliciesResult contains the result of the ingest activity.
type IngestStorageBucketIamPoliciesResult struct {
	ProjectID      string
	PolicyCount    int
	DurationMillis int64
}

// IngestStorageBucketIamPoliciesActivity is the activity function reference for workflow registration.
var IngestStorageBucketIamPoliciesActivity = (*Activities).IngestStorageBucketIamPolicies

// IngestStorageBucketIamPolicies is a Temporal activity that ingests GCP Storage bucket IAM policies.
func (a *Activities) IngestStorageBucketIamPolicies(ctx context.Context, params IngestStorageBucketIamPoliciesParams) (*IngestStorageBucketIamPoliciesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Storage bucket IAM policy ingestion",
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
		return nil, fmt.Errorf("failed to ingest bucket IAM policies: %w", err)
	}

	if err := service.DeleteStalePolicies(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale bucket IAM policies", "error", err)
	}

	logger.Info("Completed GCP Storage bucket IAM policy ingestion",
		"projectID", params.ProjectID,
		"policyCount", result.PolicyCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestStorageBucketIamPoliciesResult{
		ProjectID:      result.ProjectID,
		PolicyCount:    result.PolicyCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
