package computer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"go.temporal.io/sdk/activity"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/base/temporalerr"
	"github.com/dannyota/hotpot/pkg/ingest/meec"
	entinventory "github.com/dannyota/hotpot/pkg/storage/ent/meec/inventory"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entinventory.Client
	limiter       ratelimit.Limiter
	tokenSource   *meec.TokenSource
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entinventory.Client, limiter ratelimit.Limiter, tokenSource *meec.TokenSource) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		limiter:       limiter,
		tokenSource:   tokenSource,
	}
}

func (a *Activities) createClient() (*Client, error) {
	baseTransport := http.DefaultTransport.(*http.Transport).Clone()
	if !a.configService.MEECVerifySSL() {
		baseTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	transport := ratelimit.NewRateLimitedTransport(a.limiter, baseTransport)
	httpClient := &http.Client{Transport: transport}

	token, err := a.tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("meec authenticate: %w", err)
	}

	return NewClient(
		a.configService.MEECBaseURL(),
		token,
		a.configService.MEECAPIVersion(),
		httpClient,
	), nil
}

// IngestComputersResult contains the result of the ingest activity.
type IngestComputersResult struct {
	ComputerCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// IngestComputersActivity is the activity function reference for workflow registration.
var IngestComputersActivity = (*Activities).IngestComputers

// IngestComputers is a Temporal activity that ingests MEEC computers.
func (a *Activities) IngestComputers(ctx context.Context) (*IngestComputersResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting MEEC computer ingestion")

	client, err := a.createClient()
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(err)
	}
	service := NewService(client, a.entClient)

	result, err := service.Ingest(ctx, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest computers: %w", err))
	}

	logger.Info("Completed MEEC computer ingestion",
		"computerCount", result.ComputerCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestComputersResult{
		ComputerCount:  result.ComputerCount,
		CollectedAt:    result.CollectedAt,
		DurationMillis: result.DurationMillis,
	}, nil
}

// DeleteStaleComputersInput is the input for the DeleteStaleComputers activity.
type DeleteStaleComputersInput struct {
	CollectedAt time.Time
}

// DeleteStaleComputersActivity is the activity function reference for workflow registration.
var DeleteStaleComputersActivity = (*Activities).DeleteStaleComputers

// DeleteStaleComputers removes computers not collected in the latest run.
func (a *Activities) DeleteStaleComputers(ctx context.Context, input DeleteStaleComputersInput) error {
	client, err := a.createClient()
	if err != nil {
		return temporalerr.MaybeNonRetryable(err)
	}
	service := NewService(client, a.entClient)

	if err := service.DeleteStale(ctx, input.CollectedAt); err != nil {
		return fmt.Errorf("delete stale computers: %w", err)
	}

	return nil
}
