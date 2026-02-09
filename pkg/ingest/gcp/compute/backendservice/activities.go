package backendservice

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

// IngestComputeBackendServicesParams contains parameters for the ingest activity.
type IngestComputeBackendServicesParams struct {
	ProjectID string
}

// IngestComputeBackendServicesResult contains the result of the ingest activity.
type IngestComputeBackendServicesResult struct {
	ProjectID           string
	BackendServiceCount int
	DurationMillis      int64
}

// IngestComputeBackendServicesActivity is the activity function reference for workflow registration.
var IngestComputeBackendServicesActivity = (*Activities).IngestComputeBackendServices

// IngestComputeBackendServices is a Temporal activity that ingests GCP Compute backend services.
func (a *Activities) IngestComputeBackendServices(ctx context.Context, params IngestComputeBackendServicesParams) (*IngestComputeBackendServicesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute backend service ingestion",
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
		return nil, fmt.Errorf("failed to ingest backend services: %w", err)
	}

	// Delete stale backend services
	if err := service.DeleteStaleBackendServices(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale backend services", "error", err)
	}

	logger.Info("Completed GCP Compute backend service ingestion",
		"projectID", params.ProjectID,
		"backendServiceCount", result.BackendServiceCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeBackendServicesResult{
		ProjectID:           result.ProjectID,
		BackendServiceCount: result.BackendServiceCount,
		DurationMillis:      result.DurationMillis,
	}, nil
}
