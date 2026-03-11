package app_inventory

import (
	"context"
	"fmt"
	"net/http"

	"go.temporal.io/sdk/activity"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	ents1 "danny.vn/hotpot/pkg/storage/ent/s1"
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
		a.configService.S1BatchSize(),
		httpClient,
	)
}

// IngestS1AppInventoryResult contains the result of the ingest activity.
type IngestS1AppInventoryResult struct {
	AppCount       int
	DurationMillis int64
}

// IngestS1AppInventoryActivity is the activity function reference for workflow registration.
var IngestS1AppInventoryActivity = (*Activities).IngestS1AppInventory

// IngestS1AppInventory is a Temporal activity that ingests SentinelOne application inventory.
func (a *Activities) IngestS1AppInventory(ctx context.Context) (*IngestS1AppInventoryResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting SentinelOne app inventory ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.Ingest(ctx, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest app inventory: %w", err))
	}

	logger.Info("Completed SentinelOne app inventory ingestion",
		"appCount", result.AppCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestS1AppInventoryResult{
		AppCount:       result.AppCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
