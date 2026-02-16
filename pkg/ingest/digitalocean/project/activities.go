package project

import (
	"context"
	"fmt"
	"net/http"

	"github.com/digitalocean/godo"
	"go.temporal.io/sdk/activity"
	"golang.org/x/oauth2"

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
	rateLimitedTransport := ratelimit.NewRateLimitedTransport(a.limiter, nil)
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: a.configService.DOAPIToken()})
	oauthTransport := &oauth2.Transport{Source: tokenSource, Base: rateLimitedTransport}
	httpClient := &http.Client{Transport: oauthTransport}
	godoClient := godo.NewClient(httpClient)
	return NewClient(godoClient)
}

// IngestDOProjectsResult contains the result of the projects ingest activity.
type IngestDOProjectsResult struct {
	ProjectCount   int
	ProjectIDs     []string
	DurationMillis int64
}

// IngestDOProjectsActivity is the activity function reference for workflow registration.
var IngestDOProjectsActivity = (*Activities).IngestDOProjects

// IngestDOProjects is a Temporal activity that ingests DigitalOcean Projects.
func (a *Activities) IngestDOProjects(ctx context.Context) (*IngestDOProjectsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting DigitalOcean Project ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.IngestProjects(ctx, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("ingest projects: %w", err)
	}

	if err := service.DeleteStaleProjects(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale projects", "error", err)
	}

	logger.Info("Completed DigitalOcean Project ingestion",
		"projectCount", result.ProjectCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestDOProjectsResult{
		ProjectCount:   result.ProjectCount,
		ProjectIDs:     result.ProjectIDs,
		DurationMillis: result.DurationMillis,
	}, nil
}

// IngestDOProjectResourcesInput contains the input for the project resources ingest activity.
type IngestDOProjectResourcesInput struct {
	ProjectIDs []string
}

// IngestDOProjectResourcesResult contains the result of the project resources ingest activity.
type IngestDOProjectResourcesResult struct {
	ResourceCount  int
	DurationMillis int64
}

// IngestDOProjectResourcesActivity is the activity function reference for workflow registration.
var IngestDOProjectResourcesActivity = (*Activities).IngestDOProjectResources

// IngestDOProjectResources is a Temporal activity that ingests DigitalOcean Project Resources.
func (a *Activities) IngestDOProjectResources(ctx context.Context, input IngestDOProjectResourcesInput) (*IngestDOProjectResourcesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting DigitalOcean Project Resource ingestion", "projectCount", len(input.ProjectIDs))

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.IngestResources(ctx, input.ProjectIDs, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("ingest project resources: %w", err)
	}

	if err := service.DeleteStaleResources(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale project resources", "error", err)
	}

	logger.Info("Completed DigitalOcean Project Resource ingestion",
		"resourceCount", result.ResourceCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestDOProjectResourcesResult{
		ResourceCount:  result.ResourceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
