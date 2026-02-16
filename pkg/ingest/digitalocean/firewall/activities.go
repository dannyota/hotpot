package firewall

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

// IngestDOFirewallsResult contains the result of the ingest activity.
type IngestDOFirewallsResult struct {
	FirewallCount  int
	DurationMillis int64
}

// IngestDOFirewallsActivity is the activity function reference for workflow registration.
var IngestDOFirewallsActivity = (*Activities).IngestDOFirewalls

// IngestDOFirewalls is a Temporal activity that ingests DigitalOcean Firewalls.
func (a *Activities) IngestDOFirewalls(ctx context.Context) (*IngestDOFirewallsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting DigitalOcean Firewall ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.Ingest(ctx, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("ingest firewalls: %w", err)
	}

	if err := service.DeleteStale(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale firewalls", "error", err)
	}

	logger.Info("Completed DigitalOcean Firewall ingestion",
		"firewallCount", result.FirewallCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestDOFirewallsResult{
		FirewallCount:  result.FirewallCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
