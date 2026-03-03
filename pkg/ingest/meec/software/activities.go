package software

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

// IngestSoftwareResult contains the result of the ingest activity.
type IngestSoftwareResult struct {
	SoftwareCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// IngestSoftwareActivity is the activity function reference for workflow registration.
var IngestSoftwareActivity = (*Activities).IngestSoftware

// IngestSoftware is a Temporal activity that ingests MEEC software catalog.
func (a *Activities) IngestSoftware(ctx context.Context) (*IngestSoftwareResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting MEEC software ingestion")

	client, err := a.createClient()
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(err)
	}
	service := NewService(client, a.entClient)

	result, err := service.Ingest(ctx, func() {
		activity.RecordHeartbeat(ctx, nil)
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest software: %w", err))
	}

	logger.Info("Completed MEEC software ingestion",
		"softwareCount", result.SoftwareCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestSoftwareResult{
		SoftwareCount:  result.SoftwareCount,
		CollectedAt:    result.CollectedAt,
		DurationMillis: result.DurationMillis,
	}, nil
}

