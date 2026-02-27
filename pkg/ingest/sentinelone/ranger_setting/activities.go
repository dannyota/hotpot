package ranger_setting

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

// IngestS1RangerSettingsResult contains the result of the ingest activity.
type IngestS1RangerSettingsResult struct {
	SettingCount   int
	DurationMillis int64
}

// IngestS1RangerSettingsActivity is the activity function reference for workflow registration.
var IngestS1RangerSettingsActivity = (*Activities).IngestS1RangerSettings

// IngestS1RangerSettings is a Temporal activity that ingests SentinelOne Ranger settings.
func (a *Activities) IngestS1RangerSettings(ctx context.Context) (*IngestS1RangerSettingsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting SentinelOne Ranger setting ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.Ingest(ctx, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest ranger settings: %w", err))
	}

	if err := service.DeleteStale(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale ranger settings", "error", err)
	}

	logger.Info("Completed SentinelOne Ranger setting ingestion",
		"settingCount", result.SettingCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestS1RangerSettingsResult{
		SettingCount:   result.SettingCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
