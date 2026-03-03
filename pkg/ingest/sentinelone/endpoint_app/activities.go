package endpoint_app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"go.temporal.io/sdk/activity"
	"golang.org/x/sync/errgroup"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/base/temporalerr"
	ents1 "github.com/dannyota/hotpot/pkg/storage/ent/s1"
	"github.com/dannyota/hotpot/pkg/storage/ent/s1/bronzes1agent"
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

// ListAgentIDsResult contains the result of listing agent IDs.
type ListAgentIDsResult struct {
	AgentIDs    []string
	CollectedAt time.Time
}

// ListAgentIDsActivity is the activity function reference for workflow registration.
var ListAgentIDsActivity = (*Activities).ListAgentIDs

// ListAgentIDs queries the database for all S1 agent IDs.
func (a *Activities) ListAgentIDs(ctx context.Context) (*ListAgentIDsResult, error) {
	collectedAt := time.Now()

	agents, err := a.entClient.BronzeS1Agent.Query().
		Select(bronzes1agent.FieldID).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query agent IDs: %w", err)
	}

	agentIDs := make([]string, len(agents))
	for i, agent := range agents {
		agentIDs[i] = agent.ID
	}

	slog.Info("s1 endpoint apps: listed agent IDs", "agentCount", len(agentIDs))

	return &ListAgentIDsResult{
		AgentIDs:    agentIDs,
		CollectedAt: collectedAt,
	}, nil
}

// FetchAndSaveBatchInput is the input for the FetchAndSaveBatch activity.
type FetchAndSaveBatchInput struct {
	AgentIDs    []string
	CollectedAt time.Time
}

// FetchAndSaveBatchResult contains the result of processing a batch of agents.
type FetchAndSaveBatchResult struct {
	AppCount int
}

// FetchAndSaveBatchActivity is the activity function reference for workflow registration.
var FetchAndSaveBatchActivity = (*Activities).FetchAndSaveBatch

const fetchWorkers = 10

// FetchAndSaveBatch fetches and saves endpoint apps for a batch of agents.
// Processes up to fetchWorkers agents in parallel — the rate limiter gates throughput.
func (a *Activities) FetchAndSaveBatch(ctx context.Context, input FetchAndSaveBatchInput) (*FetchAndSaveBatchResult, error) {
	client := a.createClient()
	service := NewService(client, a.entClient)

	var totalApps atomic.Int64
	var done atomic.Int64

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(fetchWorkers)

	for _, agentID := range input.AgentIDs {
		g.Go(func() error {
			apiApps, err := client.GetEndpointApps(agentID)
			if err != nil {
				return temporalerr.MaybeNonRetryable(fmt.Errorf("get endpoint apps for agent %s: %w", agentID, err))
			}

			apps := make([]*EndpointAppData, 0, len(apiApps))
			for _, app := range apiApps {
				apps = append(apps, ConvertEndpointApp(agentID, app, input.CollectedAt))
			}

			if err := service.SaveAgentApps(gCtx, agentID, apps); err != nil {
				return fmt.Errorf("save endpoint apps for agent %s: %w", agentID, err)
			}

			totalApps.Add(int64(len(apps)))
			n := done.Add(1)
			activity.RecordHeartbeat(ctx, fmt.Sprintf("%d/%d agents, %d apps", n, len(input.AgentIDs), totalApps.Load()))
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	count := int(totalApps.Load())
	slog.Info("s1 endpoint apps: batch saved",
		"agentCount", len(input.AgentIDs),
		"appCount", count,
	)

	return &FetchAndSaveBatchResult{
		AppCount: count,
	}, nil
}

// DeleteOrphanEndpointAppsActivity is the activity function reference for workflow registration.
var DeleteOrphanEndpointAppsActivity = (*Activities).DeleteOrphanEndpointApps

// DeleteOrphanEndpointApps removes endpoint apps whose agent no longer exists.
func (a *Activities) DeleteOrphanEndpointApps(ctx context.Context) error {
	service := NewService(nil, a.entClient)

	if err := service.DeleteOrphans(ctx); err != nil {
		return fmt.Errorf("delete orphan endpoint apps: %w", err)
	}

	return nil
}
