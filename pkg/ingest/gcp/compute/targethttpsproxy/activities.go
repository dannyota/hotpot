package targethttpsproxy

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

// IngestComputeTargetHttpsProxiesParams contains parameters for the ingest activity.
type IngestComputeTargetHttpsProxiesParams struct {
	ProjectID string
}

// IngestComputeTargetHttpsProxiesResult contains the result of the ingest activity.
type IngestComputeTargetHttpsProxiesResult struct {
	ProjectID             string
	TargetHttpsProxyCount int
	DurationMillis        int64
}

// IngestComputeTargetHttpsProxiesActivity is the activity function reference for workflow registration.
var IngestComputeTargetHttpsProxiesActivity = (*Activities).IngestComputeTargetHttpsProxies

// IngestComputeTargetHttpsProxies is a Temporal activity that ingests GCP Compute target HTTPS proxies.
func (a *Activities) IngestComputeTargetHttpsProxies(ctx context.Context, params IngestComputeTargetHttpsProxiesParams) (*IngestComputeTargetHttpsProxiesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute target HTTPS proxy ingestion",
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
		return nil, fmt.Errorf("failed to ingest target HTTPS proxies: %w", err)
	}

	if err := service.DeleteStaleTargetHttpsProxies(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale target HTTPS proxies", "error", err)
	}

	logger.Info("Completed GCP Compute target HTTPS proxy ingestion",
		"projectID", params.ProjectID,
		"targetHttpsProxyCount", result.TargetHttpsProxyCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeTargetHttpsProxiesResult{
		ProjectID:             result.ProjectID,
		TargetHttpsProxyCount: result.TargetHttpsProxyCount,
		DurationMillis:        result.DurationMillis,
	}, nil
}
