package network

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
func (a *Activities) createClient(ctx context.Context, quotaProjectID string) (*Client, error) {
	httpClient, err := gcpauth.NewHTTPClient(ctx, a.configService.GCPCredentialsJSON(), a.limiter)
	if err != nil {
		return nil, err
	}
	opts := []option.ClientOption{option.WithHTTPClient(httpClient)}
	if quotaProjectID != "" {
		opts = append(opts, option.WithQuotaProject(quotaProjectID))
	}
	return NewClient(ctx, opts...)
}

// IngestComputeNetworksParams contains parameters for the ingest activity.
type IngestComputeNetworksParams struct {
	ProjectID      string
	QuotaProjectID string
}

// IngestComputeNetworksResult contains the result of the ingest activity.
type IngestComputeNetworksResult struct {
	ProjectID      string
	NetworkCount   int
	DurationMillis int64
}

// IngestComputeNetworksActivity is the activity function reference for workflow registration.
var IngestComputeNetworksActivity = (*Activities).IngestComputeNetworks

// IngestComputeNetworks is a Temporal activity that ingests GCP Compute networks.
func (a *Activities) IngestComputeNetworks(ctx context.Context, params IngestComputeNetworksParams) (*IngestComputeNetworksResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute network ingestion",
		"projectID", params.ProjectID,
	)

	// Create client for this activity
	client, err := a.createClient(ctx, params.QuotaProjectID)
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
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest networks: %w", err))
	}

	// Delete stale networks
	if err := service.DeleteStaleNetworks(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale networks", "error", err)
	}

	logger.Info("Completed GCP Compute network ingestion",
		"projectID", params.ProjectID,
		"networkCount", result.NetworkCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeNetworksResult{
		ProjectID:      result.ProjectID,
		NetworkCount:   result.NetworkCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
