package ubuntu

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

// IngestUbuntuPackagesResult contains the result of the Ubuntu packages ingest activity.
type IngestUbuntuPackagesResult struct {
	PackageCount   int
	DurationMillis int64
}

// IngestUbuntuPackagesActivity is the activity function reference for workflow registration.
var IngestUbuntuPackagesActivity = (*Activities).IngestUbuntuPackages

// IngestUbuntuPackages downloads and ingests Ubuntu package indexes.
func (a *Activities) IngestUbuntuPackages(ctx context.Context) (*IngestUbuntuPackagesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting Ubuntu packages ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.Ingest(ctx, func(details string) {
		activity.RecordHeartbeat(ctx, details)
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest Ubuntu packages: %w", err))
	}

	logger.Info("Completed Ubuntu packages ingestion",
		"packageCount", result.PackageCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestUbuntuPackagesResult{
		PackageCount:   result.PackageCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
