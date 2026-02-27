package endpoint_app

import (
	"context"
	"fmt"
	"net/http"

	"go.temporal.io/sdk/activity"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/base/temporalerr"
	ents1 "github.com/dannyota/hotpot/pkg/storage/ent/s1"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *ents1.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *ents1.Client, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		limiter:       limiter,
	}
}

func (a *Activities) createClient() *Client {
	httpClient := &http.Client{
		Transport: ratelimit.NewRateLimitedTransport(a.limiter, nil),
	}
	return NewClient(
		a.configService.S1BaseURL(),
		a.configService.S1APIToken(),
		httpClient,
	)
}

// IngestS1EndpointAppsResult contains the result of the ingest activity.
type IngestS1EndpointAppsResult struct {
	AppCount       int
	DurationMillis int64
}

// IngestS1EndpointAppsActivity is the activity function reference for workflow registration.
var IngestS1EndpointAppsActivity = (*Activities).IngestS1EndpointApps

// IngestS1EndpointApps is a Temporal activity that ingests SentinelOne endpoint applications.
func (a *Activities) IngestS1EndpointApps(ctx context.Context) (*IngestS1EndpointAppsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting SentinelOne endpoint app ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.Ingest(ctx, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest endpoint apps: %w", err))
	}

	if err := service.DeleteStale(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale endpoint apps", "error", err)
	}

	logger.Info("Completed SentinelOne endpoint app ingestion",
		"appCount", result.AppCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestS1EndpointAppsResult{
		AppCount:       result.AppCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
