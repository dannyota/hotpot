package subnetwork

import (
	"context"
	"fmt"
	"net/http"

	"go.temporal.io/sdk/activity"
	"google.golang.org/api/option"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
	"hotpot/pkg/storage/ent"
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

// IngestComputeSubnetworksParams contains parameters for the ingest activity.
type IngestComputeSubnetworksParams struct {
	ProjectID string
}

// IngestComputeSubnetworksResult contains the result of the ingest activity.
type IngestComputeSubnetworksResult struct {
	ProjectID       string
	SubnetworkCount int
	DurationMillis  int64
}

// IngestComputeSubnetworksActivity is the activity function reference for workflow registration.
var IngestComputeSubnetworksActivity = (*Activities).IngestComputeSubnetworks

// IngestComputeSubnetworks is a Temporal activity that ingests GCP Compute subnetworks.
func (a *Activities) IngestComputeSubnetworks(ctx context.Context, params IngestComputeSubnetworksParams) (*IngestComputeSubnetworksResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute subnetwork ingestion",
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
		return nil, fmt.Errorf("failed to ingest subnetworks: %w", err)
	}

	// Delete stale subnetworks
	if err := service.DeleteStaleSubnetworks(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale subnetworks", "error", err)
	}

	logger.Info("Completed GCP Compute subnetwork ingestion",
		"projectID", params.ProjectID,
		"subnetworkCount", result.SubnetworkCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeSubnetworksResult{
		ProjectID:       result.ProjectID,
		SubnetworkCount: result.SubnetworkCount,
		DurationMillis:  result.DurationMillis,
	}, nil
}

