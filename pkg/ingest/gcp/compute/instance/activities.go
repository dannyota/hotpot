package instance

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

// IngestComputeInstancesParams contains parameters for the ingest activity.
type IngestComputeInstancesParams struct {
	ProjectID string
}

// IngestComputeInstancesResult contains the result of the ingest activity.
type IngestComputeInstancesResult struct {
	ProjectID      string
	InstanceCount  int
	DurationMillis int64
}

// IngestComputeInstancesActivity is the activity function reference for workflow registration.
var IngestComputeInstancesActivity = (*Activities).IngestComputeInstances

// IngestComputeInstances is a Temporal activity that ingests GCP Compute instances.
func (a *Activities) IngestComputeInstances(ctx context.Context, params IngestComputeInstancesParams) (*IngestComputeInstancesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute instance ingestion",
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
		return nil, fmt.Errorf("failed to ingest instances: %w", err)
	}

	// Delete stale instances
	if err := service.DeleteStaleInstances(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale instances", "error", err)
	}

	logger.Info("Completed GCP Compute instance ingestion",
		"projectID", params.ProjectID,
		"instanceCount", result.InstanceCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeInstancesResult{
		ProjectID:      result.ProjectID,
		InstanceCount:  result.InstanceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}

