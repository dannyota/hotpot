package targetinstance

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

// IngestComputeTargetInstancesParams contains parameters for the ingest activity.
type IngestComputeTargetInstancesParams struct {
	ProjectID string
}

// IngestComputeTargetInstancesResult contains the result of the ingest activity.
type IngestComputeTargetInstancesResult struct {
	ProjectID           string
	TargetInstanceCount int
	DurationMillis      int64
}

// IngestComputeTargetInstancesActivity is the activity function reference for workflow registration.
var IngestComputeTargetInstancesActivity = (*Activities).IngestComputeTargetInstances

// IngestComputeTargetInstances is a Temporal activity that ingests GCP Compute target instances.
func (a *Activities) IngestComputeTargetInstances(ctx context.Context, params IngestComputeTargetInstancesParams) (*IngestComputeTargetInstancesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute target instance ingestion",
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
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest target instances: %w", err))
	}

	// Delete stale target instances
	if err := service.DeleteStaleTargetInstances(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale target instances", "error", err)
	}

	logger.Info("Completed GCP Compute target instance ingestion",
		"projectID", params.ProjectID,
		"targetInstanceCount", result.TargetInstanceCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeTargetInstancesResult{
		ProjectID:           result.ProjectID,
		TargetInstanceCount: result.TargetInstanceCount,
		DurationMillis:      result.DurationMillis,
	}, nil
}
