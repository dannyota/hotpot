package packetmirroring

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

// IngestComputePacketMirroringsParams contains parameters for the ingest activity.
type IngestComputePacketMirroringsParams struct {
	ProjectID      string
	QuotaProjectID string
}

// IngestComputePacketMirroringsResult contains the result of the ingest activity.
type IngestComputePacketMirroringsResult struct {
	ProjectID            string
	PacketMirroringCount int
	DurationMillis       int64
}

// IngestComputePacketMirroringsActivity is the activity function reference for workflow registration.
var IngestComputePacketMirroringsActivity = (*Activities).IngestComputePacketMirrorings

// IngestComputePacketMirrorings is a Temporal activity that ingests GCP Compute packet mirrorings.
func (a *Activities) IngestComputePacketMirrorings(ctx context.Context, params IngestComputePacketMirroringsParams) (*IngestComputePacketMirroringsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute packet mirroring ingestion",
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
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("failed to ingest packet mirrorings: %w", err))
	}

	// Delete stale packet mirrorings
	if err := service.DeleteStalePacketMirrorings(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale packet mirrorings", "error", err)
	}

	logger.Info("Completed GCP Compute packet mirroring ingestion",
		"projectID", params.ProjectID,
		"packetMirroringCount", result.PacketMirroringCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputePacketMirroringsResult{
		ProjectID:            result.ProjectID,
		PacketMirroringCount: result.PacketMirroringCount,
		DurationMillis:       result.DurationMillis,
	}, nil
}
