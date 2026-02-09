package agent

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

// IngestS1AgentsResult contains the result of the ingest activity.
type IngestS1AgentsResult struct {
	AgentCount     int
	DurationMillis int64
}

// IngestS1AgentsActivity is the activity function reference for workflow registration.
var IngestS1AgentsActivity = (*Activities).IngestS1Agents

// IngestS1Agents is a Temporal activity that ingests SentinelOne agents with cursor pagination.
func (a *Activities) IngestS1Agents(ctx context.Context) (*IngestS1AgentsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting SentinelOne agent ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.Ingest(ctx, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("ingest agents: %w", err)
	}

	if err := service.DeleteStale(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale agents", "error", err)
	}

	logger.Info("Completed SentinelOne agent ingestion",
		"agentCount", result.AgentCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestS1AgentsResult{
		AgentCount:     result.AgentCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
