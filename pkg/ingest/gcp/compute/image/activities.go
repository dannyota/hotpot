package image

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
	var opts []option.ClientOption
	opts = append(opts, option.WithHTTPClient(httpClient))
	if quotaProjectID != "" {
		opts = append(opts, option.WithQuotaProject(quotaProjectID))
	}
	return NewClient(ctx, opts...)
}

// IngestComputeImagesParams contains parameters for the ingest activity.
type IngestComputeImagesParams struct {
	ProjectID      string
	QuotaProjectID string
}

// IngestComputeImagesResult contains the result of the ingest activity.
type IngestComputeImagesResult struct {
	ProjectID      string
	ImageCount     int
	DurationMillis int64
}

// IngestComputeImagesActivity is the activity function reference for workflow registration.
var IngestComputeImagesActivity = (*Activities).IngestComputeImages

// IngestComputeImages is a Temporal activity that ingests GCP Compute images.
func (a *Activities) IngestComputeImages(ctx context.Context, params IngestComputeImagesParams) (*IngestComputeImagesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute image ingestion",
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
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest images: %w", err))
	}

	// Delete stale images
	if err := service.DeleteStaleImages(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale images", "error", err)
	}

	logger.Info("Completed GCP Compute image ingestion",
		"projectID", params.ProjectID,
		"imageCount", result.ImageCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeImagesResult{
		ProjectID:      result.ProjectID,
		ImageCount:     result.ImageCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
