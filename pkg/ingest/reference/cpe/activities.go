package cpe

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

// IngestCPEResult contains the result of the CPE ingest activity.
type IngestCPEResult struct {
	CPECount       int
	DurationMillis int64
}

// IngestCPEActivity is the activity function reference for workflow registration.
var IngestCPEActivity = (*Activities).IngestCPE

// IngestCPE downloads and ingests the NVD CPE Dictionary.
func (a *Activities) IngestCPE(ctx context.Context) (*IngestCPEResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting CPE ingestion")

	client := a.createClient()
	service := NewService(client, a.entClient)

	result, err := service.Ingest(ctx, func(details string) {
		activity.RecordHeartbeat(ctx, details)
	})
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("ingest CPE: %w", err))
	}

	logger.Info("Completed CPE ingestion",
		"cpeCount", result.CPECount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestCPEResult{
		CPECount:       result.CPECount,
		DurationMillis: result.DurationMillis,
	}, nil
}
