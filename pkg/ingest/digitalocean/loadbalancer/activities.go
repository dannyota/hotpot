package loadbalancer

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

// IngestDOLoadBalancersResult contains the result of the ingest activity.
type IngestDOLoadBalancersResult struct {
	LoadBalancerCount int
	DurationMillis    int64
}

// IngestDOLoadBalancersActivity is the activity function reference for workflow registration.
var IngestDOLoadBalancersActivity = (*Activities).IngestDOLoadBalancers

// IngestDOLoadBalancers is a Temporal activity that ingests DigitalOcean Load Balancers.
func (a *Activities) IngestDOLoadBalancers(ctx context.Context) (*IngestDOLoadBalancersResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting DigitalOcean Load Balancer ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.Ingest(ctx, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("ingest load balancers: %w", err)
	}

	if err := service.DeleteStale(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale load balancers", "error", err)
	}

	logger.Info("Completed DigitalOcean Load Balancer ingestion",
		"loadBalancerCount", result.LoadBalancerCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestDOLoadBalancersResult{
		LoadBalancerCount: result.LoadBalancerCount,
		DurationMillis:    result.DurationMillis,
	}, nil
}
