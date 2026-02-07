package ingest

import (
	"context"
	"fmt"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
	"hotpot/pkg/ingest/gcp"
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
func Run(ctx context.Context, configService *config.Service, db *gorm.DB) error {
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
	rateLimitSvc := gcp.Register(gcpWorker, configService, db)
	defer rateLimitSvc.Close()

	// Create and register VNGCloud worker (future)
	// vngWorker := worker.New(temporalClient, TaskQueueVNGCloud, worker.Options{})
	// vngcloud.Register(vngWorker, cfg.VNGCloud, db)

	// Create and register SentinelOne worker (future)
	// s1Worker := worker.New(temporalClient, TaskQueueS1, worker.Options{})
	// sentinelone.Register(s1Worker, cfg.S1, db)

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

	// Run GCP worker (blocking until interrupted)
	err = gcpWorker.Run(interruptCh)
	if err != nil {
		return fmt.Errorf("GCP worker failed: %w", err)
	}

	return nil
}
