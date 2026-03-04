package eol

import (
	"context"
	"fmt"
	"net/http"

	"go.temporal.io/sdk/activity"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/base/temporalerr"
	entreference "github.com/dannyota/hotpot/pkg/storage/ent/reference"
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

// IngestEOLResult contains the result of the EOL ingest activity.
type IngestEOLResult struct {
	ProductCount   int
	CycleCount     int
	DurationMillis int64
}

// IngestEOLActivity is the activity function reference for workflow registration.
var IngestEOLActivity = (*Activities).IngestEOL

// IngestEOL downloads and ingests the endoflife.date database.
func (a *Activities) IngestEOL(ctx context.Context) (*IngestEOLResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting EOL ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.Ingest(ctx, func(details string) {
		activity.RecordHeartbeat(ctx, details)
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest EOL: %w", err))
	}

	logger.Info("Completed EOL ingestion",
		"productCount", result.ProductCount,
		"cycleCount", result.CycleCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestEOLResult{
		ProductCount:   result.ProductCount,
		CycleCount:     result.CycleCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
