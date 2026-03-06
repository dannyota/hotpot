package ubuntu

import (
	"context"
	"fmt"
	"net/http"

	"go.temporal.io/sdk/activity"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	entreference "danny.vn/hotpot/pkg/storage/ent/reference"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entreference.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entreference.Client, limiter ratelimit.Limiter) *Activities {
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
	return NewClient(httpClient)
}

// IngestUbuntuFeedInput is the input for the per-feed ingest activity.
type IngestUbuntuFeedInput struct {
	Release   string
	Component string
}

// IngestUbuntuFeedResult contains the result of a single Ubuntu feed ingest activity.
type IngestUbuntuFeedResult struct {
	Release        string
	Component      string
	PackageCount   int
	DurationMillis int64
}

// IngestUbuntuFeedActivity is the activity function reference for workflow registration.
var IngestUbuntuFeedActivity = (*Activities).IngestUbuntuFeed

// IngestUbuntuFeed downloads and ingests a single Ubuntu Packages.gz feed.
func (a *Activities) IngestUbuntuFeed(ctx context.Context, input IngestUbuntuFeedInput) (*IngestUbuntuFeedResult, error) {
	logger := activity.GetLogger(ctx)
	label := input.Release + "/" + input.Component
	logger.Info("Starting Ubuntu feed ingestion", "feed", label)

	// Find feed definition
	var feed FeedDef
	var found bool
	for _, f := range Feeds {
		if f.Release == input.Release && f.Component == input.Component {
			feed = f
			found = true
			break
		}
	}
	if !found {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("unknown Ubuntu feed: %s", label))
	}

	heartbeat := func(details string) {
		activity.RecordHeartbeat(ctx, details)
	}

	client := a.createClient()
	packages, err := client.DownloadFeed(feed, heartbeat)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("download Ubuntu feed %s: %w", label, err))
	}

	service := NewService(a.entClient)
	result, err := service.IngestFeed(ctx, input.Release, input.Component, packages, heartbeat)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest Ubuntu feed %s: %w", label, err))
	}

	logger.Info("Completed Ubuntu feed ingestion",
		"feed", label,
		"packageCount", result.PackageCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestUbuntuFeedResult{
		Release:        result.Release,
		Component:      result.Component,
		PackageCount:   result.PackageCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
