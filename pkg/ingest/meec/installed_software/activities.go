package installed_software

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"go.temporal.io/sdk/activity"
	"golang.org/x/sync/errgroup"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/base/temporalerr"
	"danny.vn/hotpot/pkg/ingest/meec"
	entinventory "danny.vn/hotpot/pkg/storage/ent/meec/inventory"
	"danny.vn/hotpot/pkg/storage/ent/meec/inventory/bronzemeecinventorycomputer"
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

// ListComputerIDsResult contains the result of listing computer IDs.
type ListComputerIDsResult struct {
	ComputerIDs []string
	CollectedAt time.Time
}

// ListComputerIDsActivity is the activity function reference for workflow registration.
var ListComputerIDsActivity = (*Activities).ListComputerIDs

// ListComputerIDs queries the database for all MEEC computer IDs.
func (a *Activities) ListComputerIDs(ctx context.Context) (*ListComputerIDsResult, error) {
	collectedAt := time.Now()

	computers, err := a.entClient.BronzeMEECInventoryComputer.Query().
		Select(bronzemeecinventorycomputer.FieldID).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query computer IDs: %w", err)
	}

	computerIDs := make([]string, len(computers))
	for i, c := range computers {
		computerIDs[i] = c.ID
	}

	slog.Info("meec installed software: listed computer IDs", "computerCount", len(computerIDs))

	return &ListComputerIDsResult{
		ComputerIDs: computerIDs,
		CollectedAt: collectedAt,
	}, nil
}

// FetchAndSaveBatchInput is the input for the FetchAndSaveBatch activity.
type FetchAndSaveBatchInput struct {
	ComputerIDs []string
	CollectedAt time.Time
}

// FetchAndSaveBatchResult contains the result of processing a batch of computers.
type FetchAndSaveBatchResult struct {
	SoftwareCount int
}

// FetchAndSaveBatchActivity is the activity function reference for workflow registration.
var FetchAndSaveBatchActivity = (*Activities).FetchAndSaveBatch

const fetchWorkers = 10

// FetchAndSaveBatch fetches and saves installed software for a batch of computers.
// Processes up to fetchWorkers computers in parallel — the rate limiter gates throughput.
func (a *Activities) FetchAndSaveBatch(ctx context.Context, input FetchAndSaveBatchInput) (*FetchAndSaveBatchResult, error) {
	client, err := a.createClient()
	if err != nil {
		return nil, temporalerr.MaybeNonRetryable(err)
	}
	service := NewService(client, a.entClient)

	var totalSoftware atomic.Int64
	var done atomic.Int64

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(fetchWorkers)

	for _, computerID := range input.ComputerIDs {
		g.Go(func() error {
			apiSoftware, err := client.GetInstalledSoftware(computerID)
			if err != nil {
				return temporalerr.MaybeNonRetryable(fmt.Errorf("get installed software for computer %s: %w", computerID, err))
			}

			software := make([]*InstalledSoftwareData, 0, len(apiSoftware))
			for _, s := range apiSoftware {
				software = append(software, ConvertInstalledSoftware(computerID, s, input.CollectedAt))
			}

			if err := service.SaveComputerSoftware(gCtx, computerID, software); err != nil {
				return fmt.Errorf("save installed software for computer %s: %w", computerID, err)
			}

			totalSoftware.Add(int64(len(software)))
			n := done.Add(1)
			activity.RecordHeartbeat(ctx, fmt.Sprintf("%d/%d computers, %d software", n, len(input.ComputerIDs), totalSoftware.Load()))
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	count := int(totalSoftware.Load())
	slog.Info("meec installed software: batch saved",
		"computerCount", len(input.ComputerIDs),
		"softwareCount", count,
	)

	return &FetchAndSaveBatchResult{
		SoftwareCount: count,
	}, nil
}

// DeleteOrphanInstalledSoftwareActivity is the activity function reference for workflow registration.
var DeleteOrphanInstalledSoftwareActivity = (*Activities).DeleteOrphanInstalledSoftware

// DeleteOrphanInstalledSoftware removes installed software whose computer no longer exists.
func (a *Activities) DeleteOrphanInstalledSoftware(ctx context.Context) error {
	service := NewService(nil, a.entClient)

	if err := service.DeleteOrphans(ctx); err != nil {
		return fmt.Errorf("delete orphan installed software: %w", err)
	}

	return nil
}
