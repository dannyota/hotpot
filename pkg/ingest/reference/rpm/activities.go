package rpm

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

// IngestRPMPackagesResult contains the result of the RPM packages ingest activity.
type IngestRPMPackagesResult struct {
	PackageCount   int
	DurationMillis int64
}

// IngestRPMPackagesActivity is the activity function reference for workflow registration.
var IngestRPMPackagesActivity = (*Activities).IngestRPMPackages

// IngestRPMPackages downloads and ingests RPM repository metadata.
func (a *Activities) IngestRPMPackages(ctx context.Context) (*IngestRPMPackagesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting RPM packages ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.Ingest(ctx, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest RPM packages: %w", err))
	}

	logger.Info("Completed RPM packages ingestion",
		"packageCount", result.PackageCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestRPMPackagesResult{
		PackageCount:   result.PackageCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
