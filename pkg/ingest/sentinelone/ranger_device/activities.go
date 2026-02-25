package ranger_device

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go.temporal.io/sdk/activity"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
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

// IngestS1RangerDevicesResult contains the result of the ingest activity.
type IngestS1RangerDevicesResult struct {
	DeviceCount    int
	DurationMillis int64
}

// IngestS1RangerDevicesActivity is the activity function reference for workflow registration.
var IngestS1RangerDevicesActivity = (*Activities).IngestS1RangerDevices

// IngestS1RangerDevices is a Temporal activity that ingests SentinelOne ranger devices with cursor pagination.
func (a *Activities) IngestS1RangerDevices(ctx context.Context) (*IngestS1RangerDevicesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting SentinelOne ranger device ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.Ingest(ctx, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		if strings.Contains(err.Error(), "status 403") || strings.Contains(err.Error(), "authentication failed") {
			logger.Warn("Ranger not licensed, skipping device ingestion", "error", err)
			return &IngestS1RangerDevicesResult{DeviceCount: 0}, nil
		}
		return nil, fmt.Errorf("ingest ranger devices: %w", err)
	}

	if err := service.DeleteStale(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale ranger devices", "error", err)
	}

	logger.Info("Completed SentinelOne ranger device ingestion",
		"deviceCount", result.DeviceCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestS1RangerDevicesResult{
		DeviceCount:    result.DeviceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
