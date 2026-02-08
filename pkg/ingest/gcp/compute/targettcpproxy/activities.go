package targettcpproxy

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

// IngestComputeTargetTcpProxiesParams contains parameters for the ingest activity.
type IngestComputeTargetTcpProxiesParams struct {
	ProjectID string
}

// IngestComputeTargetTcpProxiesResult contains the result of the ingest activity.
type IngestComputeTargetTcpProxiesResult struct {
	ProjectID           string
	TargetTcpProxyCount int
	DurationMillis      int64
}

// IngestComputeTargetTcpProxiesActivity is the activity function reference for workflow registration.
var IngestComputeTargetTcpProxiesActivity = (*Activities).IngestComputeTargetTcpProxies

// IngestComputeTargetTcpProxies is a Temporal activity that ingests GCP Compute target TCP proxies.
func (a *Activities) IngestComputeTargetTcpProxies(ctx context.Context, params IngestComputeTargetTcpProxiesParams) (*IngestComputeTargetTcpProxiesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute target TCP proxy ingestion",
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
		return nil, fmt.Errorf("failed to ingest target TCP proxies: %w", err)
	}

	if err := service.DeleteStaleTargetTcpProxies(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale target TCP proxies", "error", err)
	}

	logger.Info("Completed GCP Compute target TCP proxy ingestion",
		"projectID", params.ProjectID,
		"targetTcpProxyCount", result.TargetTcpProxyCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeTargetTcpProxiesResult{
		ProjectID:           result.ProjectID,
		TargetTcpProxyCount: result.TargetTcpProxyCount,
		DurationMillis:      result.DurationMillis,
	}, nil
}
