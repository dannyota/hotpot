package eol

import (
	"context"
	"fmt"
	"net/http"
	"time"

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

// IngestEOLResult contains the result of the EOL ingest activity.
type IngestEOLResult struct {
	ProductCount    int
	CycleCount      int
	IdentifierCount int
	DurationMillis  int64
}

// IngestEOLActivity is the activity function reference for workflow registration.
var IngestEOLActivity = (*Activities).IngestEOL

// IngestRHELEUSActivity is the activity function reference for workflow registration.
var IngestRHELEUSActivity = (*Activities).IngestRHELEUS

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
		"identifierCount", result.IdentifierCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestEOLResult{
		ProductCount:    result.ProductCount,
		CycleCount:      result.CycleCount,
		IdentifierCount: result.IdentifierCount,
		DurationMillis:  result.DurationMillis,
	}, nil
}

// IngestRHELEUSResult contains the result of the RHEL EUS ingest activity.
type IngestRHELEUSResult struct {
	CycleCount     int
	DurationMillis int64
}

// IngestRHELEUS fetches RHEL EUS data from Red Hat and inserts cycles into the EOL table.
func (a *Activities) IngestRHELEUS(ctx context.Context) (*IngestRHELEUSResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting RHEL EUS ingestion")

	start := time.Now()

	httpClient := &http.Client{
		Transport: ratelimit.NewRateLimitedTransport(a.limiter, nil),
	}

	activity.RecordHeartbeat(ctx, "fetching Red Hat errata page")

	data, err := FetchRHELEUS(httpClient)
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(fmt.Errorf("fetch RHEL EUS: %w", err))
	}

	logger.Info("Parsed RHEL EUS data", "eus", len(data.EUSCycles), "enhanced", len(data.EnhancedEUSCycles))
	activity.RecordHeartbeat(ctx, fmt.Sprintf("parsed %d EUS + %d Enhanced EUS cycles", len(data.EUSCycles), len(data.EnhancedEUSCycles)))

	// Build a map of cycle -> EUS end date and Enhanced EUS end date.
	type eusDates struct {
		eol  *time.Time // Standard EUS end date
		eoes *time.Time // Enhanced EUS end date
	}
	cycleMap := make(map[string]*eusDates)
	for _, c := range data.EUSCycles {
		if _, ok := cycleMap[c.Cycle]; !ok {
			cycleMap[c.Cycle] = &eusDates{}
		}
		cycleMap[c.Cycle].eol = c.EndDate
	}
	for _, c := range data.EnhancedEUSCycles {
		if _, ok := cycleMap[c.Cycle]; !ok {
			cycleMap[c.Cycle] = &eusDates{}
		}
		cycleMap[c.Cycle].eoes = c.EndDate
	}

	now := time.Now()
	count := 0

	for cycle, dates := range cycleMap {
		id := "rhel:" + cycle
		b := a.entClient.BronzeReferenceEOLCycle.Create().
			SetID(id).
			SetProduct("rhel").
			SetCycle(cycle).
			SetLatest(cycle).
			SetCollectedAt(now).
			SetFirstCollectedAt(now)

		if dates.eol != nil {
			b.SetEol(*dates.eol)
		}
		if dates.eoes != nil {
			b.SetEoes(*dates.eoes)
		}

		if err := b.Exec(ctx); err != nil {
			return nil, fmt.Errorf("insert RHEL EUS cycle %s: %w", cycle, err)
		}
		count++
	}

	logger.Info("Completed RHEL EUS ingestion", "cycles", count, "durationMillis", time.Since(start).Milliseconds())

	return &IngestRHELEUSResult{
		CycleCount:     count,
		DurationMillis: time.Since(start).Milliseconds(),
	}, nil
}
