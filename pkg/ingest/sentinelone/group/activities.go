package group

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
		maxGroupsBatchSize,
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
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest groups: %w", err))
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
