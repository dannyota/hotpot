package address

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

// IngestComputeAddressesParams contains parameters for the ingest activity.
type IngestComputeAddressesParams struct {
	ProjectID string
}

// IngestComputeAddressesResult contains the result of the ingest activity.
type IngestComputeAddressesResult struct {
	ProjectID      string
	AddressCount   int
	DurationMillis int64
}

// IngestComputeAddressesActivity is the activity function reference for workflow registration.
var IngestComputeAddressesActivity = (*Activities).IngestComputeAddresses

// IngestComputeAddresses is a Temporal activity that ingests GCP Compute regional addresses.
func (a *Activities) IngestComputeAddresses(ctx context.Context, params IngestComputeAddressesParams) (*IngestComputeAddressesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute address ingestion",
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
		return nil, fmt.Errorf("failed to ingest addresses: %w", err)
	}

	// Delete stale addresses
	if err := service.DeleteStaleAddresses(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale addresses", "error", err)
	}

	logger.Info("Completed GCP Compute address ingestion",
		"projectID", params.ProjectID,
		"addressCount", result.AddressCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeAddressesResult{
		ProjectID:      result.ProjectID,
		AddressCount:   result.AddressCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
