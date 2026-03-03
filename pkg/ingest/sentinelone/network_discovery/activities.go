package network_discovery

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
		a.configService.S1BatchSize(),
		httpClient,
	)
}

// IngestS1NetworkDiscoveryResult contains the result of the ingest activity.
type IngestS1NetworkDiscoveryResult struct {
	DeviceCount    int
	DurationMillis int64
}

// IngestS1NetworkDiscoveryActivity is the activity function reference for workflow registration.
var IngestS1NetworkDiscoveryActivity = (*Activities).IngestS1NetworkDiscovery

// IngestS1NetworkDiscovery is a Temporal activity that ingests SentinelOne network discovery devices with cursor pagination.
func (a *Activities) IngestS1NetworkDiscovery(ctx context.Context) (*IngestS1NetworkDiscoveryResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting SentinelOne network discovery ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.Ingest(ctx, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest network discovery devices: %w", err))
	}

	logger.Info("Completed SentinelOne network discovery ingestion",
		"deviceCount", result.DeviceCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestS1NetworkDiscoveryResult{
		DeviceCount:    result.DeviceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
