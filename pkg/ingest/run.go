package ingest

import (
	"context"
	"fmt"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"golang.org/x/sync/errgroup"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
	"hotpot/pkg/ingest/gcp"
	"hotpot/pkg/ingest/sentinelone"
	"hotpot/pkg/storage/ent"
)

// TaskQueues for different providers.
const (
	TaskQueueGCP      = "hotpot-ingest-gcp"
	TaskQueueVNGCloud = "hotpot-ingest-vng"
	TaskQueueS1       = "hotpot-ingest-s1"
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

	// Server-side activity rate limit: always set as a safety net.
	// Caps activity dispatches across all workers on this task queue.
	reqPerMin := configService.GCPRateLimitPerMinute()
	activitiesPerSec := float64(reqPerMin) / 60.0

	// Create GCP worker
	gcpWorker := worker.New(temporalClient, TaskQueueGCP, worker.Options{
		TaskQueueActivitiesPerSecond: activitiesPerSec,
	})

	// Register GCP workflows and activities
	rateLimitSvc := gcp.Register(gcpWorker, configService, entClient)
	defer rateLimitSvc.Close()

	// Create and register VNGCloud worker (future)
	// vngWorker := worker.New(temporalClient, TaskQueueVNGCloud, worker.Options{})
	// vngcloud.Register(vngWorker, cfg.VNGCloud, db)

	// Create and register SentinelOne worker if configured
	var s1RateLimitSvc *ratelimit.Service
	var s1Worker worker.Worker
	if configService.S1Configured() {
		s1ReqPerMin := configService.S1RateLimitPerMinute()
		s1ActivitiesPerSec := float64(s1ReqPerMin) / 60.0

		s1Worker = worker.New(temporalClient, TaskQueueS1, worker.Options{
			TaskQueueActivitiesPerSecond: s1ActivitiesPerSec,
		})

		s1RateLimitSvc = sentinelone.Register(s1Worker, configService, entClient)
		defer s1RateLimitSvc.Close()

		log.Printf("SentinelOne worker configured (rate limit: %d rpm, Temporal: %.1f activities/sec)",
			s1ReqPerMin, s1ActivitiesPerSec)
	}

	// Create and register Fortinet worker (future)
	// fortinetWorker := worker.New(temporalClient, TaskQueueFortinet, worker.Options{})
	// fortinet.Register(fortinetWorker, cfg.Fortinet, db)

	log.Printf("Starting ingest workers (GCP rate limit: %d rpm, Temporal: %.1f activities/sec)...",
		reqPerMin, activitiesPerSec)

	// Convert context cancellation to interrupt channel for Temporal worker
	interruptCh := make(chan interface{})
	go func() {
		<-ctx.Done()
		close(interruptCh)
	}()

	// Run workers concurrently
	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := gcpWorker.Run(interruptCh); err != nil {
			return fmt.Errorf("GCP worker failed: %w", err)
		}
		return nil
	})

	if s1Worker != nil {
		g.Go(func() error {
			if err := s1Worker.Run(interruptCh); err != nil {
				return fmt.Errorf("S1 worker failed: %w", err)
			}
			return nil
		})
	}

	return g.Wait()
}
