package xeol

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

// IngestXeolResult contains the result of the xeol ingest activity.
type IngestXeolResult struct {
	ProductCount   int
	CycleCount     int
	PurlCount      int
	VulnCount      int
	DurationMillis int64
}

// IngestXeolActivity is the activity function reference for workflow registration.
var IngestXeolActivity = (*Activities).IngestXeol

// IngestXeol downloads and ingests the xeol EOL database.
func (a *Activities) IngestXeol(ctx context.Context) (*IngestXeolResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting xeol ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.Ingest(ctx, func(details string) {
		activity.RecordHeartbeat(ctx, details)
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest xeol: %w", err))
	}

	logger.Info("Completed xeol ingestion",
		"productCount", result.ProductCount,
		"cycleCount", result.CycleCount,
		"purlCount", result.PurlCount,
		"vulnCount", result.VulnCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestXeolResult{
		ProductCount:   result.ProductCount,
		CycleCount:     result.CycleCount,
		PurlCount:      result.PurlCount,
		VulnCount:      result.VulnCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
