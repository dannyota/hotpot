package endpoint_app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"go.temporal.io/sdk/activity"

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
	logger := activity.GetLogger(ctx)
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

	logger.Info("Listed S1 agent IDs", "agentCount", len(agentIDs))

	return &ListAgentIDsResult{
		AgentIDs:    agentIDs,
		CollectedAt: collectedAt,
	}, nil
}

// FetchAndSaveAgentAppsInput is the input for the FetchAndSaveAgentApps activity.
type FetchAndSaveAgentAppsInput struct {
	AgentID     string
	CollectedAt time.Time
}

// FetchAndSaveAgentAppsResult contains the result of fetching and saving apps for one agent.
type FetchAndSaveAgentAppsResult struct {
	AppCount int
}

// FetchAndSaveAgentAppsActivity is the activity function reference for workflow registration.
var FetchAndSaveAgentAppsActivity = (*Activities).FetchAndSaveAgentApps

// FetchAndSaveAgentApps fetches endpoint apps for a single agent and saves them.
func (a *Activities) FetchAndSaveAgentApps(ctx context.Context, input FetchAndSaveAgentAppsInput) (*FetchAndSaveAgentAppsResult, error) {
	client := a.createClient()

	apiApps, err := client.GetEndpointApps(input.AgentID)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("get endpoint apps for agent %s: %w", input.AgentID, err))
	}

	apps := make([]*EndpointAppData, 0, len(apiApps))
	for _, app := range apiApps {
		apps = append(apps, ConvertEndpointApp(input.AgentID, app, input.CollectedAt))
	}

	activity.RecordHeartbeat(ctx, input.AgentID)

	service := NewService(client, a.entClient)
	if err := service.SaveAgentApps(ctx, apps); err != nil {
		return nil, fmt.Errorf("save endpoint apps for agent %s: %w", input.AgentID, err)
	}

	slog.Debug("s1 endpoint apps: saved agent apps", "agentID", input.AgentID, "appCount", len(apps))

	return &FetchAndSaveAgentAppsResult{
		AppCount: len(apps),
	}, nil
}

// DeleteStaleEndpointAppsInput is the input for the DeleteStaleEndpointApps activity.
type DeleteStaleEndpointAppsInput struct {
	CollectedAt time.Time
}

// DeleteStaleEndpointAppsActivity is the activity function reference for workflow registration.
var DeleteStaleEndpointAppsActivity = (*Activities).DeleteStaleEndpointApps

// DeleteStaleEndpointApps removes endpoint apps not collected in the latest run.
func (a *Activities) DeleteStaleEndpointApps(ctx context.Context, input DeleteStaleEndpointAppsInput) error {
	service := NewService(a.createClient(), a.entClient)

	if err := service.DeleteStale(ctx, input.CollectedAt); err != nil {
		return fmt.Errorf("delete stale endpoint apps: %w", err)
	}

	return nil
}
