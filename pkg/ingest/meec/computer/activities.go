package computer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"go.temporal.io/sdk/activity"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	"danny.vn/hotpot/pkg/ingest/meec"
	entinventory "danny.vn/hotpot/pkg/storage/ent/meec/inventory"
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

