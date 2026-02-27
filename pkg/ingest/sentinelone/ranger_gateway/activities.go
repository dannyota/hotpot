package ranger_gateway

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

// IngestS1RangerGatewaysResult contains the result of the ingest activity.
type IngestS1RangerGatewaysResult struct {
	GatewayCount   int
	DurationMillis int64
}

// IngestS1RangerGatewaysActivity is the activity function reference for workflow registration.
var IngestS1RangerGatewaysActivity = (*Activities).IngestS1RangerGateways

// IngestS1RangerGateways is a Temporal activity that ingests SentinelOne ranger gateways with cursor pagination.
func (a *Activities) IngestS1RangerGateways(ctx context.Context) (*IngestS1RangerGatewaysResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting SentinelOne ranger gateway ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.Ingest(ctx, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest ranger gateways: %w", err))
	}

	if err := service.DeleteStale(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale ranger gateways", "error", err)
	}

	logger.Info("Completed SentinelOne ranger gateway ingestion",
		"gatewayCount", result.GatewayCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestS1RangerGatewaysResult{
		GatewayCount:   result.GatewayCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
