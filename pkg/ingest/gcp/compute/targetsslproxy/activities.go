package targetsslproxy

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

// IngestComputeTargetSslProxiesParams contains parameters for the ingest activity.
type IngestComputeTargetSslProxiesParams struct {
	ProjectID string
}

// IngestComputeTargetSslProxiesResult contains the result of the ingest activity.
type IngestComputeTargetSslProxiesResult struct {
	ProjectID           string
	TargetSslProxyCount int
	DurationMillis      int64
}

// IngestComputeTargetSslProxiesActivity is the activity function reference.
var IngestComputeTargetSslProxiesActivity = (*Activities).IngestComputeTargetSslProxies

// IngestComputeTargetSslProxies is a Temporal activity that ingests GCP Compute target SSL proxies.
func (a *Activities) IngestComputeTargetSslProxies(ctx context.Context, params IngestComputeTargetSslProxiesParams) (*IngestComputeTargetSslProxiesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting GCP Compute target SSL proxy ingestion", "projectID", params.ProjectID)

	client, err := a.createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{ProjectID: params.ProjectID})
	if err != nil {
		return nil, fmt.Errorf("failed to ingest target SSL proxies: %w", err)
	}

	if err := service.DeleteStaleTargetSslProxies(ctx, params.ProjectID, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale target SSL proxies", "error", err)
	}

	logger.Info("Completed GCP Compute target SSL proxy ingestion",
		"projectID", params.ProjectID,
		"targetSslProxyCount", result.TargetSslProxyCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputeTargetSslProxiesResult{
		ProjectID:           result.ProjectID,
		TargetSslProxyCount: result.TargetSslProxyCount,
		DurationMillis:      result.DurationMillis,
	}, nil
}
