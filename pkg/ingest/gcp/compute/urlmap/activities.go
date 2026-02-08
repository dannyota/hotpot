package urlmap

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

// IngestComputeUrlMapsParams contains parameters for the ingest activity.
type IngestComputeUrlMapsParams struct {
	ProjectID string
}

// IngestComputeUrlMapsResult contains the result of the ingest activity.
type IngestComputeUrlMapsResult struct {
	ProjectID      string
	UrlMapCount    int
	DurationMillis int64
}

// IngestComputeUrlMapsActivity is the activity function reference for workflow registration.
var IngestComputeUrlMapsActivity = (*Activities).IngestComputeUrlMaps

// IngestComputeUrlMaps is a Temporal activity that ingests GCP Compute URL maps.
func (a *Activities) IngestComputeUrlMaps(ctx context.Context, params IngestComputeUrlMapsParams) (*IngestComputeUrlMapsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute URL map ingestion",
		"projectID", params.ProjectID,
	)

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{
		ProjectID: params.ProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest URL maps: %w", err)
	}

	if err := service.DeleteStaleUrlMaps(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale URL maps", "error", err)
	}

	logger.Info("Completed GCP Compute URL map ingestion",
		"projectID", params.ProjectID,
		"urlMapCount", result.UrlMapCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeUrlMapsResult{
		ProjectID:      result.ProjectID,
		UrlMapCount:    result.UrlMapCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
