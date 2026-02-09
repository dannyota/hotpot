package threat

import (
	"context"
	"fmt"
	"net/http"

	"go.temporal.io/sdk/activity"

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

func (a *Activities) createClient() *Client {
	httpClient := &http.Client{
		Transport: ratelimit.NewRateLimitedTransport(a.limiter, nil),
	}
	return NewClient(
		a.configService.S1BaseURL(),
		a.configService.S1APIToken(),
		a.configService.S1BatchSize(),
		httpClient,
	)
}

// IngestS1ThreatsResult contains the result of the ingest activity.
type IngestS1ThreatsResult struct {
	ThreatCount    int
	DurationMillis int64
}

// IngestS1ThreatsActivity is the activity function reference for workflow registration.
var IngestS1ThreatsActivity = (*Activities).IngestS1Threats

// IngestS1Threats is a Temporal activity that ingests SentinelOne threats with cursor pagination.
func (a *Activities) IngestS1Threats(ctx context.Context) (*IngestS1ThreatsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting SentinelOne threat ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.Ingest(ctx, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("ingest threats: %w", err)
	}

	if err := service.DeleteStale(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale threats", "error", err)
	}

	logger.Info("Completed SentinelOne threat ingestion",
		"threatCount", result.ThreatCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestS1ThreatsResult{
		ThreatCount:    result.ThreatCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
