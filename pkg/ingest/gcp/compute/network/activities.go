package network

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

// IngestComputeNetworksParams contains parameters for the ingest activity.
type IngestComputeNetworksParams struct {
	ProjectID string
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
		return nil, fmt.Errorf("failed to ingest networks: %w", err)
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

