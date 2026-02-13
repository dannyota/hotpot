package ingest

import (
	"context"
	"fmt"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"golang.org/x/sync/errgroup"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
	"github.com/dannyota/hotpot/pkg/ingest/sentinelone"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// TaskQueues for different providers.
const (
	TaskQueueGCP      = "hotpot-ingest-gcp"
	TaskQueueVNGCloud = "hotpot-ingest-vng"
	TaskQueueS1       = "hotpot-ingest-s1"
	TaskQueueDO       = "hotpot-ingest-do"
	TaskQueueFortinet = "hotpot-ingest-fortinet"
)

// Run starts the ingest workers.
// The context is used to signal shutdown - when cancelled, workers will stop.
func Run(ctx context.Context, configService *config.Service, entClient *ent.Client) error {
	// Create Temporal client
	temporalClient, err := client.Dial(client.Options{
		HostPort:  configService.TemporalHostPort(),
		Namespace: configService.TemporalNamespace(),
	})
	if err != nil {
		return fmt.Errorf("failed to create Temporal client: %w", err)
	}
	defer temporalClient.Close()

	enabled := configService.EnabledProviders()
	if len(enabled) == 0 {
		return fmt.Errorf("no providers enabled in config; set enabled: true for at least one provider")
	}
	log.Printf("Enabled providers: %v", enabled)

	// Convert context cancellation to interrupt channel for Temporal worker
	interruptCh := make(chan interface{})
	go func() {
		<-ctx.Done()
		close(interruptCh)
	}()

	// Run workers concurrently
	g, _ := errgroup.WithContext(ctx)

	// GCP
	if configService.GCPEnabled() {
		reqPerMin := configService.GCPRateLimitPerMinute()
		activitiesPerSec := float64(reqPerMin) / 60.0

		gcpWorker := worker.New(temporalClient, TaskQueueGCP, worker.Options{
			TaskQueueActivitiesPerSecond: activitiesPerSec,
		})

		rateLimitSvc := gcp.Register(gcpWorker, configService, entClient)
		defer rateLimitSvc.Close()

		log.Printf("GCP worker enabled (rate limit: %d rpm, Temporal: %.1f activities/sec)",
			reqPerMin, activitiesPerSec)

		g.Go(func() error {
			if err := gcpWorker.Run(interruptCh); err != nil {
				return fmt.Errorf("GCP worker failed: %w", err)
			}
			return nil
		})
	}

	// SentinelOne
	if configService.S1Enabled() {
		s1ReqPerMin := configService.S1RateLimitPerMinute()
		s1ActivitiesPerSec := float64(s1ReqPerMin) / 60.0

		s1Worker := worker.New(temporalClient, TaskQueueS1, worker.Options{
			TaskQueueActivitiesPerSecond: s1ActivitiesPerSec,
		})

		s1RateLimitSvc := sentinelone.Register(s1Worker, configService, entClient)
		defer s1RateLimitSvc.Close()

		log.Printf("SentinelOne worker enabled (rate limit: %d rpm, Temporal: %.1f activities/sec)",
			s1ReqPerMin, s1ActivitiesPerSec)

		g.Go(func() error {
			if err := s1Worker.Run(interruptCh); err != nil {
				return fmt.Errorf("S1 worker failed: %w", err)
			}
			return nil
		})
	}

	// DigitalOcean
	if configService.DOEnabled() {
		doReqPerMin := configService.DORateLimitPerMinute()
		doActivitiesPerSec := float64(doReqPerMin) / 60.0

		doWorker := worker.New(temporalClient, TaskQueueDO, worker.Options{
			TaskQueueActivitiesPerSecond: doActivitiesPerSec,
		})

		doRateLimitSvc := digitalocean.Register(doWorker, configService, entClient)
		defer doRateLimitSvc.Close()

		log.Printf("DigitalOcean worker enabled (rate limit: %d rpm, Temporal: %.1f activities/sec)",
			doReqPerMin, doActivitiesPerSec)

		g.Go(func() error {
			if err := doWorker.Run(interruptCh); err != nil {
				return fmt.Errorf("DO worker failed: %w", err)
			}
			return nil
		})
	}

	return g.Wait()
}
