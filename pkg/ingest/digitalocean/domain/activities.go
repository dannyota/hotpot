package domain

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

// IngestDODomainsResult contains the result of the domains ingest activity.
type IngestDODomainsResult struct {
	DomainCount    int
	DomainNames    []string
	DurationMillis int64
}

// IngestDODomainsActivity is the activity function reference for workflow registration.
var IngestDODomainsActivity = (*Activities).IngestDODomains

// IngestDODomains is a Temporal activity that ingests DigitalOcean Domains.
func (a *Activities) IngestDODomains(ctx context.Context) (*IngestDODomainsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting DigitalOcean Domain ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.IngestDomains(ctx, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("ingest domains: %w", err)
	}

	if err := service.DeleteStaleDomains(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale domains", "error", err)
	}

	logger.Info("Completed DigitalOcean Domain ingestion",
		"domainCount", result.DomainCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestDODomainsResult{
		DomainCount:    result.DomainCount,
		DomainNames:    result.DomainNames,
		DurationMillis: result.DurationMillis,
	}, nil
}

// IngestDODomainRecordsInput contains the input for the domain records ingest activity.
type IngestDODomainRecordsInput struct {
	DomainNames []string
}

// IngestDODomainRecordsResult contains the result of the domain records ingest activity.
type IngestDODomainRecordsResult struct {
	RecordCount    int
	DurationMillis int64
}

// IngestDODomainRecordsActivity is the activity function reference for workflow registration.
var IngestDODomainRecordsActivity = (*Activities).IngestDODomainRecords

// IngestDODomainRecords is a Temporal activity that ingests DigitalOcean Domain Records.
func (a *Activities) IngestDODomainRecords(ctx context.Context, input IngestDODomainRecordsInput) (*IngestDODomainRecordsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting DigitalOcean Domain Record ingestion", "domainCount", len(input.DomainNames))

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.IngestRecords(ctx, input.DomainNames, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("ingest domain records: %w", err)
	}

	if err := service.DeleteStaleRecords(ctx, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale domain records", "error", err)
	}

	logger.Info("Completed DigitalOcean Domain Record ingestion",
		"recordCount", result.RecordCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestDODomainRecordsResult{
		RecordCount:    result.RecordCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
