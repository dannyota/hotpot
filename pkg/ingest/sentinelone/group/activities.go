package group

import (
	"context"
	"fmt"
	"net/http"

	"go.temporal.io/sdk/activity"

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

// IngestS1GroupsResult contains the result of the ingest activity.
type IngestS1GroupsResult struct {
	GroupCount     int
	DurationMillis int64
}

// IngestS1GroupsActivity is the activity function reference for workflow registration.
var IngestS1GroupsActivity = (*Activities).IngestS1Groups

// IngestS1Groups is a Temporal activity that ingests SentinelOne groups.
func (a *Activities) IngestS1Groups(ctx context.Context) (*IngestS1GroupsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting SentinelOne group ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.Ingest(ctx, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("ingest groups: %w", err)
	}

	if err := service.DeleteStale(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale groups", "error", err)
	}

	logger.Info("Completed SentinelOne group ingestion",
		"groupCount", result.GroupCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestS1GroupsResult{
		GroupCount:     result.GroupCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
