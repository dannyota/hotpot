package site

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

// IngestS1SitesResult contains the result of the ingest activity.
type IngestS1SitesResult struct {
	SiteCount      int
	DurationMillis int64
}

// IngestS1SitesActivity is the activity function reference for workflow registration.
var IngestS1SitesActivity = (*Activities).IngestS1Sites

// IngestS1Sites is a Temporal activity that ingests SentinelOne sites.
func (a *Activities) IngestS1Sites(ctx context.Context) (*IngestS1SitesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting SentinelOne site ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.Ingest(ctx, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("ingest sites: %w", err)
	}

	if err := service.DeleteStale(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale sites", "error", err)
	}

	logger.Info("Completed SentinelOne site ingestion",
		"siteCount", result.SiteCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestS1SitesResult{
		SiteCount:      result.SiteCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
