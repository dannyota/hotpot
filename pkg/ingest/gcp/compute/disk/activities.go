package disk

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

// IngestComputeDisksParams contains parameters for the ingest activity.
type IngestComputeDisksParams struct {
	ProjectID      string
	QuotaProjectID string
}

// IngestComputeDisksResult contains the result of the ingest activity.
type IngestComputeDisksResult struct {
	ProjectID      string
	DiskCount      int
	DurationMillis int64
}

// IngestComputeDisksActivity is the activity function reference for workflow registration.
var IngestComputeDisksActivity = (*Activities).IngestComputeDisks

// IngestComputeDisks is a Temporal activity that ingests GCP Compute disks.
func (a *Activities) IngestComputeDisks(ctx context.Context, params IngestComputeDisksParams) (*IngestComputeDisksResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute disk ingestion",
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
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest disks: %w", err))
	}

	// Delete stale disks
	if err := service.DeleteStaleDisks(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale disks", "error", err)
	}

	logger.Info("Completed GCP Compute disk ingestion",
		"projectID", params.ProjectID,
		"diskCount", result.DiskCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeDisksResult{
		ProjectID:      result.ProjectID,
		DiskCount:      result.DiskCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
